package service

type GitHubClient interface {
	RepositoryExists(owner, repo string) (bool, error)
	GetLatestReleaseTag(owner, repo string) (*string, error)
}

type MailSender interface {
	SendConfirmationEmail(email, confirmToken, unsubscribeToken, repo string) error
	SendNewReleaseEmail(email, repo, tag, unsubscribeToken string) error
}
