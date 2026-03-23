package service

import (
	"bytes"
	"context"
	"core/config"
	"core/domain"
	"core/models"
	"core/queue"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/pgvector/pgvector-go"
)

type GitHubRepositoryService struct {
	GitHubRepositoryDomain    domain.GitHubRepositoryDomain
	GitHubCommitsDomain       domain.GitHubCommitsDomain
	GitHubCommitFilesDomain   domain.GitHubCommitFilesDomain
	CommitFileEmbeddingDomain domain.CommitFileEmbeddingDomain
	QueueClient               *queue.Client
}

// BackfillEmbeddings queues embedding tasks for every commit file that has no
// embedding yet.  Returns the number of tasks enqueued.
func (g *GitHubRepositoryService) BackfillEmbeddings() (int, error) {
	if g.QueueClient == nil {
		return 0, fmt.Errorf("queue client not configured")
	}

	files, err := g.GitHubCommitFilesDomain.GetUnembeddedCommitFiles()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch unembedded files: %w", err)
	}

	enqueued := 0
	for _, f := range files {
		// Apply the same extension filter used during normal ingestion
		if !isEmbeddableFile(f.Filename) {
			continue
		}
		if err := g.QueueClient.EnqueueEmbedCommitFile(queue.EmbedCommitFilePayload{
			RepoFullName:  f.FullName,
			RepoGithubID:  f.GithubRepoID,
			RepoInstallID: f.InstallationID,
			CommitID:      f.GithubCommitID,
			CommitSHA:     "",
			CommitMessage: f.Message,
			AuthorName:    f.Author,
			FileID:        f.CommitFileID,
			Filename:      f.Filename,
			Patch:         f.Patch,
		}); err != nil {
			fmt.Printf("[backfill] failed to enqueue file_id=%d: %v\n", f.CommitFileID, err)
			continue
		}
		enqueued++
	}

	fmt.Printf("[backfill] enqueued %d / %d unembedded files\n", enqueued, len(files))
	return enqueued, nil
}

func (g *GitHubRepositoryService) GetRepositoryActivity(repoID, days int64) ([]models.CommitActivity, error) {
	data, err := g.GitHubRepositoryDomain.GetRepositoryActivity(repoID, days)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (g *GitHubRepositoryService) GetCommitDetails(repoID int64, commitSHA string) (models.CommitDetailsResponse, error) {
	data, err := g.GitHubCommitsDomain.GetCommitDetails(repoID, commitSHA)
	if err != nil {
		return models.CommitDetailsResponse{}, err
	}
	return data, nil
}

func (g *GitHubRepositoryService) GetRelatedCommitFiles(commitFileID string) ([]models.RelatedCommitFileResponse, error) {
	data, err := g.CommitFileEmbeddingDomain.GetRelatedCommitFiles(commitFileID)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (g *GitHubRepositoryService) ExplainCommitFileChange(param models.ExplainCommitFileChangeRequest) (models.ExplainCommitFileChangeResponse, error) {
	data, err := g.CommitFileEmbeddingDomain.GetRelatedCommitFiles(param.CommitFileID)
	if err != nil {
		return models.ExplainCommitFileChangeResponse{}, err
	}

	commitFileInt, err := strconv.ParseInt(param.CommitFileID, 10, 64)
	if err != nil {
		return models.ExplainCommitFileChangeResponse{}, fmt.Errorf("invalid commit file id: %w", err)
	}

	// Get the main commit file details
	mainCommitFile, err := g.GitHubCommitFilesDomain.GetGitHubCommitFilesByID(commitFileInt)
	if err != nil {
		return models.ExplainCommitFileChangeResponse{}, err
	}

	// Limit the number of related files to top 3 most similar
	maxRelatedFiles := 3
	if len(data) > maxRelatedFiles {
		data = data[:maxRelatedFiles]
	}

	// Build related files context
	relatedFilesContext, err := g.buildRelatedFilesContext(data)
	if err != nil {
		return models.ExplainCommitFileChangeResponse{}, err
	}

	// Build prompts
	systemPrompt := g.buildSystemPrompt()
	userPrompt := g.buildUserPrompt(mainCommitFile, relatedFilesContext, len(data), param.Question)

	// Log token estimation (rough: 1 token ≈ 4 chars)
	estimatedTokens := (len(systemPrompt) + len(userPrompt)) / 4
	fmt.Printf("Estimated tokens: %d\n", estimatedTokens)

	aiResponse, err := g.callAIService(systemPrompt, userPrompt)
	if err != nil {
		return models.ExplainCommitFileChangeResponse{}, fmt.Errorf("AI service error: %w", err)
	}
	response := models.ExplainCommitFileChangeResponse{
		Summary: aiResponse,
		Reasoning: []string{
			fmt.Sprintf("Main file: %s (%s)", mainCommitFile.Filename, mainCommitFile.Status),
			fmt.Sprintf("Analyzed %d related historical changes", len(data)),
		},
		ConfidenceScore: 0.9, // Placeholder confidence score
	}

	return response, nil
}

func (g *GitHubRepositoryService) callAIService(systemPrompt, userPrompt string) (string, error) {
	aiServiceURL := config.GetConfig().AiBackendUrl + "/explain-commit-file-change"

	// Prepare request payload
	requestBody := map[string]string{
		"systemPrompt": systemPrompt,
		"userPrompt":   userPrompt,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP POST request
	resp, err := http.Post(aiServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Return the response as string (assuming the AI service returns plain text)
	return string(body), nil
}

func (g *GitHubRepositoryService) buildRelatedFilesContext(relatedFiles []models.RelatedCommitFileResponse) (string, error) {
	const (
		maxPatchChars        = 500  // Limit patch size per file
		maxTotalContextChars = 2000 // Limit total related files context
	)

	var relatedFilesContext string
	totalContextLength := 0

	for _, file := range relatedFiles {
		githubCommitFiles, err := g.GitHubCommitFilesDomain.GetGitHubCommitFilesByID(file.CommitFileID)
		if err != nil {
			continue // Skip files with errors
		}

		// Truncate patch if too long
		patch := githubCommitFiles.Patch
		if len(patch) > maxPatchChars {
			patch = patch[:maxPatchChars] + "\n... (truncated)"
		}

		fileContext := fmt.Sprintf(`
Related File %d:
- Filename: %s
- Status: %s
- Additions: %d, Deletions: %d
- Commit SHA: %s
- Author: %s
- Similarity Score: %.2f
- Patch:
%s
---
`, file.CommitFileID, githubCommitFiles.Filename, githubCommitFiles.Status,
			githubCommitFiles.Additions, githubCommitFiles.Deletions,
			file.CommitSHA, file.Author, file.Similarity, patch)

		// Check if adding this would exceed the limit
		if totalContextLength+len(fileContext) > maxTotalContextChars {
			relatedFilesContext += "\n... (additional related files omitted)"
			break
		}

		relatedFilesContext += fileContext
		totalContextLength += len(fileContext)
	}

	return relatedFilesContext, nil
}

func (g *GitHubRepositoryService) buildSystemPrompt() string {
	return `You are an expert code reviewer. Analyze commit changes and explain:
1. What changed and why
2. How it relates to historical changes
3. Potential impact and patterns

Be concise and technical.`
}

func (g *GitHubRepositoryService) QueryWorkspace(query string, workspaceID int64) (models.WorkspaceQueryResponse, error) {
	ctx := context.Background()

	// 1. Generate query embedding
	embedding, err := GenerateEmbeddingRaw(ctx, query)
	if err != nil {
		return models.WorkspaceQueryResponse{}, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// 2. Vector similarity search scoped to workspace
	results, err := g.CommitFileEmbeddingDomain.VectorSearchByWorkspace(
		pgvector.NewVector(embedding), workspaceID, 10,
	)
	if err != nil {
		return models.WorkspaceQueryResponse{}, fmt.Errorf("vector search failed: %w", err)
	}

	if len(results) == 0 {
		return models.WorkspaceQueryResponse{
			Answer:  "No relevant code history found for this query in the workspace.",
			Sources: []models.WorkspaceQuerySource{},
		}, nil
	}

	// 3. Build context string for LLM
	context := g.buildWorkspaceContext(results)

	// 4. Build prompts
	systemPrompt := `You are an expert code analyst. Answer the user's question based on the provided commit history context.
Be concise and technical. Reference specific files and commits when relevant.`

	userPrompt := fmt.Sprintf(`Based on the following commit history from the codebase, answer this question:

QUESTION: %s

COMMIT HISTORY CONTEXT:
%s

Provide a clear, direct answer referencing the relevant changes.`, query, context)

	// 5. Call LLM
	answer, err := g.callAzureChatCompletion(systemPrompt, userPrompt)
	if err != nil {
		return models.WorkspaceQueryResponse{}, fmt.Errorf("LLM call failed: %w", err)
	}

	// 6. Build sources (deduplicated by commit_file_id)
	seen := make(map[int64]bool)
	var sources []models.WorkspaceQuerySource
	for _, r := range results {
		if seen[r.CommitFileID] {
			continue
		}
		seen[r.CommitFileID] = true
		sources = append(sources, models.WorkspaceQuerySource{
			FileName:  r.Filename,
			CommitSHA: r.CommitSHA,
			RepoName:  r.RepoName,
		})
	}

	return models.WorkspaceQueryResponse{Answer: answer, Sources: sources}, nil
}

func (g *GitHubRepositoryService) buildWorkspaceContext(results []models.WorkspaceSearchResult) string {
	const maxPatchChars = 400
	const maxTotalChars = 3000

	var sb strings.Builder
	for i, r := range results {
		patch := r.Patch
		if len(patch) > maxPatchChars {
			patch = patch[:maxPatchChars] + "\n... (truncated)"
		}
		entry := fmt.Sprintf(`[%d] File: %s | Repo: %s | Commit: %s | Author: %s | Similarity: %.2f
Patch:
%s
---
`, i+1, r.Filename, r.RepoName, r.CommitSHA, r.Author, r.Similarity, patch)

		if sb.Len()+len(entry) > maxTotalChars {
			break
		}
		sb.WriteString(entry)
	}
	return sb.String()
}

func (g *GitHubRepositoryService) callAzureChatCompletion(systemPrompt, userPrompt string) (string, error) {
	aiServiceURL := config.GetConfig().AiBackendUrl + "/explain-commit-file-change"

	requestBody := map[string]string{
		"systemPrompt": systemPrompt,
		"userPrompt":   userPrompt,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(aiServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read AI service response: %w", err)
	}

	return string(body), nil
}

func (g *GitHubRepositoryService) buildUserPrompt(mainCommitFile models.GitHubCommitFiles, relatedFilesContext string, relatedCount int, question string) string {
	const maxMainPatchChars = 1000

	// Truncate main file patch if too long
	mainPatch := mainCommitFile.Patch
	if len(mainPatch) > maxMainPatchChars {
		mainPatch = mainPatch[:maxMainPatchChars] + "\n... (truncated)"
	}

	return fmt.Sprintf(`Analyze this commit file change:

MAIN CHANGE:
File: %s (%s)
+%d/-%d lines
Patch:
%s

RELATED CHANGES (%d similar commits):
%s

Provide a brief explanation of what changed, why, and any patterns from historical changes.

QUESTION:
%s`, mainCommitFile.Filename, mainCommitFile.Status,
		mainCommitFile.Additions, mainCommitFile.Deletions, mainPatch,
		relatedCount, relatedFilesContext, question)
}
