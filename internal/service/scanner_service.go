package service

import (
	"log"
	"strings"

	"github-release-notification-api/internal/repository"
)

type ScannerService struct {
	repoRepo         repository.RepositoryRepository
	subscriptionRepo repository.SubscriptionRepository
	githubClient     GitHubClient
	mailSender       MailSender
}

func NewScannerService(
	repoRepo repository.RepositoryRepository,
	subscriptionRepo repository.SubscriptionRepository,
	githubClient GitHubClient,
	mailSender MailSender,
) *ScannerService {
	return &ScannerService{
		repoRepo:         repoRepo,
		subscriptionRepo: subscriptionRepo,
		githubClient:     githubClient,
		mailSender:       mailSender,
	}
}

func (s *ScannerService) ScanOnce() {
	repos, err := s.repoRepo.GetTrackedRepositories()
	if err != nil {
		log.Printf("scanner: failed to load tracked repositories: %v", err)
		return
	}

	for _, repo := range repos {
		tag, err := s.githubClient.GetLatestReleaseTag(repo.Owner, repo.Name)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "rate limit") {
				log.Printf("scanner: github rate limit for %s: %v", repo.FullName, err)
				continue
			}
			log.Printf("scanner: failed to get latest release for %s: %v", repo.FullName, err)
			continue
		}

		if err := s.repoRepo.UpdateLastCheckedAt(repo.ID); err != nil {
			log.Printf("scanner: failed to update last_checked_at for %s: %v", repo.FullName, err)
		}

		if tag == nil {
			continue
		}

		if repo.LastSeenTag != nil && *repo.LastSeenTag == *tag {
			continue
		}

		subscriptions, err := s.subscriptionRepo.GetActiveConfirmedByRepositoryID(repo.ID)
		if err != nil {
			log.Printf("scanner: failed to load subscriptions for %s: %v", repo.FullName, err)
			continue
		}

		sendFailed := false

		for _, subscription := range subscriptions {
			if err := s.mailSender.SendNewReleaseEmail(
				subscription.Email,
				repo.FullName,
				*tag,
				subscription.UnsubscribeToken,
			); err != nil {
				log.Printf("scanner: failed to send release email to %s: %v", subscription.Email, err)
				sendFailed = true
			}
		}

		if sendFailed {
			log.Printf("scanner: skipping last_seen_tag update for %s because some emails failed", repo.FullName)
			continue
		}

		if err := s.repoRepo.UpdateLastSeenTag(repo.ID, tag); err != nil {
			log.Printf("scanner: failed to update last_seen_tag for %s: %v", repo.FullName, err)
		}
	}
}
