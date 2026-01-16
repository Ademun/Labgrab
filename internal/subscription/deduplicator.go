package subscription

import (
	"context"
	"crypto/sha3"
	"encoding/hex"
	"fmt"
	"labgrab/pkg/config"

	"github.com/redis/go-redis/v9"
)

type Deduplicator struct {
	cache *redis.Client
	cfg   *config.DeduplicatorConfig
}

func NewDeduplicator(cache *redis.Client, cfg *config.DeduplicatorConfig) *Deduplicator {
	return &Deduplicator{cache: cache, cfg: cfg}
}

func (d *Deduplicator) Deduplicate(
	ctx context.Context,
	req *GetMatchingSubscriptionsReq,
	matches []DBSubscriptionMatchResult,
) ([]DBSubscriptionMatchResult, error) {
	var result []DBSubscriptionMatchResult

	for _, match := range matches {
		hasNewSlot := false

		for day, lessons := range match.MatchingTimeslots {
			for _, lesson := range lessons {
				key := d.generateKey(
					&keyGenerationParams{
						subscriptionUUID: match.SubscriptionUUID,
						labType:          req.LabType,
						labTopic:         req.LabTopic,
						labNumber:        req.LabNumber,
						labAuditorium:    req.LabAuditorium,
						day:              day,
						lesson:           lesson,
					},
				)

				exists, err := d.cache.Exists(ctx, key).Result()
				if err != nil {
					return nil, fmt.Errorf("failed to check key existence: %w", err)
				}

				if exists > 0 {
					err = d.cache.Expire(ctx, key, d.cfg.TTL).Err()
					if err != nil {
						return nil, fmt.Errorf("failed to update TTL: %w", err)
					}
				} else {
					hasNewSlot = true

					err = d.cache.Set(ctx, key, "1", d.cfg.TTL).Err()
					if err != nil {
						return nil, fmt.Errorf("failed to set key: %w", err)
					}
				}
			}
		}

		if hasNewSlot {
			result = append(result, match)
		}
	}

	return result, nil
}

func (d *Deduplicator) generateKey(params *keyGenerationParams) string {
	data := fmt.Sprintf("%s:%s:%d:%d:%s:%s:%d",
		params.labType,
		params.labTopic,
		params.labNumber,
		params.labAuditorium,
		params.subscriptionUUID.String(),
		params.day,
		params.lesson,
	)

	hash := sha3.New256()
	hash.Write([]byte(data))
	hashBytes := hash.Sum(nil)
	hashHex := hex.EncodeToString(hashBytes)
	key := fmt.Sprintf("%s:%s", d.cfg.KeyPrefix, hashHex)

	return key
}
