package confab

import (
	"errors"
	"time"
)

type agentRunner interface {
	Run() error
	Stop() error
}

type agentClient interface {
	VerifyJoined() error
	VerifySynced() error
	IsLastNode() (bool, error)
	SetKeys([]string) error
	Leave() error
}

type clock interface {
	Sleep(time.Duration)
}

type Controller struct {
	AgentRunner    agentRunner
	AgentClient    agentClient
	MaxRetries     int
	SyncRetryDelay time.Duration
	SyncRetryClock clock
	EncryptKeys    []string
	SSLDisabled    bool
}

func (c Controller) bootAgent() error {
	err := c.AgentRunner.Run()
	if err != nil {
		return err
	}

	for i := 1; i <= c.MaxRetries; i++ {
		err := c.AgentClient.VerifyJoined()
		if err != nil {
			if i == c.MaxRetries {
				return err
			}

			c.SyncRetryClock.Sleep(c.SyncRetryDelay)
			continue
		}

		break
	}

	return nil
}

func (c Controller) BootClient() error {
	return c.bootAgent()
}

func (c Controller) BootServer() error {
	err := c.bootAgent()
	if err != nil {
		return err
	}

	lastNode, err := c.AgentClient.IsLastNode()
	if err != nil {
		return err
	}

	if lastNode {
		for i := 1; i <= c.MaxRetries; i++ {
			err = c.AgentClient.VerifySynced()
			if err != nil {
				if i == c.MaxRetries {
					return err
				}

				c.SyncRetryClock.Sleep(c.SyncRetryDelay)
				continue
			}

			break
		}
	}

	if !c.SSLDisabled {
		if len(c.EncryptKeys) == 0 {
			return errors.New("encrypt keys cannot be empty if ssl is enabled")
		}

		err = c.AgentClient.SetKeys(c.EncryptKeys)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Controller) StopAgent() error {
	err := c.AgentClient.Leave()
	if err != nil {
		return err
	}

	err = c.AgentRunner.Stop()
	if err != nil {
		return err
	}

	return nil
}