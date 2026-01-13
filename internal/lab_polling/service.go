package lab_polling

import "labgrab/internal/shared/api/dikidi"

type Service struct {
	dikidiClient *dikidi.Client
	slotParser   *Parser
}

func NewService(client *dikidi.Client, slotParser *Parser) *Service {
	return &Service{
		dikidiClient: client,
		slotParser:   slotParser,
	}
}
