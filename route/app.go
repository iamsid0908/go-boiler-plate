package route

import (
	"core/config"
	"core/domain"
	"core/handler"
	"core/queue"
	"core/service"
	"log"

	"github.com/hibiken/asynq"
)

type AppModel struct {
	Health           handler.HealthHandler
	User             handler.UserHandler
	Auth             handler.AuthHandler
	Role             handler.RoleHandler
	Book             handler.BookHandler
	BookSummary      handler.BookSummaryHandler
	Cart             handler.CartHandler
	Workspace        handler.WorkspaceHandler
	Channel          handler.ChannelHandler
	Document         handler.DocumentHandler
	ManageWorkspace  handler.ManageWorkspaceHandler
	ManageChannels   handler.ManageChannelsHandler
	Credentials      handler.CredentialsHandler
	ConnectOrg       handler.ConnectOrgHandler
	GitHubRepository handler.GitHubRepositoryHandler
}

func App() AppModel {
	// Initialize queue client
	redisAddr := config.GetConfig().RedisAddr
	if redisAddr == "" {
		redisAddr = "localhost:6379" // default
	}
	queueClient := queue.NewClient(redisAddr)
	log.Printf("[queue] Client connected to Redis at %s", redisAddr)

	//domain
	healthDomain := &domain.HealthDomainCtx{}
	authDomain := &domain.AuthDomainCtx{}
	userDomain := &domain.UserDomainCtx{}
	roleDomain := &domain.RoleDomainCtx{}
	bookDomain := &domain.BookDomainCtx{}
	bookSummaryDomain := &domain.BookSummaryDomainCtx{}
	cartDomain := &domain.CartDomainCtx{}
	workspaceDomain := &domain.WorkspaceDomainCtx{}
	channelDomain := &domain.ChannelDomainCtx{}
	documentDomain := &domain.DocumentDomainCtx{}
	manageWorkspaceDomain := &domain.ManageWorkspaceDomainCtx{}
	manageChannelsDomain := &domain.ManageChannelsDomainCtx{}
	credentialsDomain := &domain.CredentialsDomainCtx{}
	connectOrgDomain := &domain.ConnectOrgDomainCtx{}
	gitHubCommitsDomain := &domain.GitHubCommitsDomainCtx{}
	gitHubRepositoryDomain := &domain.GitHubRepositoryDomainCtx{}
	gitHubCommitFilesDomain := &domain.GitHubCommitFilesDomainCtx{}
	commitFileEmbeddingDomain := &domain.CommitFileEmbeddingDomainCtx{}
	gitHubInstallationsDomain := &domain.GitHubInstallationsDomainCtx{}

	//service
	healthService := service.HealthService{
		HealthDomain: healthDomain,
	}
	userService := service.UserService{
		UserDomain: userDomain,
	}
	authService := service.AuthService{
		AuthDomain: authDomain,
		UserDomain: userDomain,
	}
	roleService := service.RoleService{
		RoleDomain: roleDomain,
	}
	bookService := service.BookService{
		BookDomain:        bookDomain,
		BookSummaryDomain: bookSummaryDomain,
	}
	bookSummaryService := service.BookSummaryService{
		BookSummaryDomain: bookSummaryDomain,
	}
	cartService := service.CartService{
		CartDomain: cartDomain,
	}
	workspaceService := service.WorkspaceService{
		WorkspaceDomain:           workspaceDomain,
		ManageWorkspaceDomain:     manageWorkspaceDomain,
		ChannelDomain:             channelDomain,
		DocumentDomain:            documentDomain,
		CredentialsDomain:         credentialsDomain,
		ManageChannelsDomain:      manageChannelsDomain,
		UserDomain:                userDomain,
		GitHubRepositoryDomain:    gitHubRepositoryDomain,
		GitHubInstallationsDomain: gitHubInstallationsDomain,
		GitHubCommitsDomain:       gitHubCommitsDomain,
		GitHubCommitFilesDomain:   gitHubCommitFilesDomain,
	}
	channelService := service.ChannelService{
		ChannelDomain:         channelDomain,
		ManageChannelsDomain:  manageChannelsDomain,
		UserDomain:            userDomain,
		ManageWorkspaceDomain: manageWorkspaceDomain,
	}
	documentService := service.DocumentService{
		DocumentDomain: documentDomain,
	}
	manageWorkspaceService := service.ManageWorkspaceService{
		ManageWorkspaceDomain: manageWorkspaceDomain,
	}
	manageChannelsService := service.ManageChannelsService{
		ManageChannelsDomain: manageChannelsDomain,
	}
	credentialsService := service.CredentialsService{
		CredentialsDomain: credentialsDomain,
	}
	connectOrgService := service.ConnectOrgService{
		ConnectOrgDomain:          connectOrgDomain,
		GitHubCommitsDomain:       gitHubCommitsDomain,
		GitHubRepositoryDomain:    gitHubRepositoryDomain,
		GitHubCommitFilesDomain:   gitHubCommitFilesDomain,
		CommitFileEmbeddingDomain: commitFileEmbeddingDomain,
		QueueClient:               queueClient,
	}

	// Start the asynq worker server (processes enqueued tasks in background)
	mux := asynq.NewServeMux()
	RegisterTaskHandlers(mux, &connectOrgService)
	queue.StartWorker(redisAddr, mux)

	gitHubRepositoryService := service.GitHubRepositoryService{
		GitHubRepositoryDomain:    gitHubRepositoryDomain,
		GitHubCommitsDomain:       gitHubCommitsDomain,
		GitHubCommitFilesDomain:   gitHubCommitFilesDomain,
		CommitFileEmbeddingDomain: commitFileEmbeddingDomain,
	}

	//handler
	healthHandler := handler.HealthHandler{
		HealthService: healthService,
	}
	userHandler := handler.UserHandler{
		UserService: userService,
	}
	authHandler := handler.AuthHandler{
		AuthService: authService,
	}
	roleHandler := handler.RoleHandler{
		RoleService: roleService,
	}
	bookHandler := handler.BookHandler{
		BookService: bookService,
	}
	bookSummaryHandler := handler.BookSummaryHandler{
		BookSummaryService: bookSummaryService,
	}
	cartHandler := handler.CartHandler{
		CartService: cartService,
	}
	workspaceHandler := handler.WorkspaceHandler{
		WorkspaceService: workspaceService,
	}
	channelHandler := handler.ChannelHandler{
		ChannelService: channelService,
	}
	documentHandler := handler.DocumentHandler{
		DocumentService: documentService,
	}
	manageWorkspaceHandler := handler.ManageWorkspaceHandler{
		ManageWorkspaceService: manageWorkspaceService,
	}
	manageChannelsHandler := handler.ManageChannelsHandler{
		ManageChannelsService: manageChannelsService,
	}
	credentialsHandler := handler.CredentialsHandler{
		CredentialsService: credentialsService,
	}
	connectOrgHandler := handler.ConnectOrgHandler{
		ConnectOrgService: connectOrgService,
	}
	gitHubRepositoryHandler := handler.GitHubRepositoryHandler{
		GitHubRepositoryService: gitHubRepositoryService,
	}

	return AppModel{
		Health:           healthHandler,
		User:             userHandler,
		Auth:             authHandler,
		Role:             roleHandler,
		Book:             bookHandler,
		BookSummary:      bookSummaryHandler,
		Cart:             cartHandler,
		Workspace:        workspaceHandler,
		Channel:          channelHandler,
		Document:         documentHandler,
		ManageWorkspace:  manageWorkspaceHandler,
		ManageChannels:   manageChannelsHandler,
		Credentials:      credentialsHandler,
		ConnectOrg:       connectOrgHandler,
		GitHubRepository: gitHubRepositoryHandler,
	}
}
