package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
)

type APICacheManager struct {
	client          *mongo.Client
	cacheCollection *mongo.Collection
	httpClient      *http.Client
	rateLimiter     *rate.Limiter
	logger          *zap.SugaredLogger
	cacheExpiry     time.Duration
}

type APIResponse struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
}

type CacheEntry struct {
	ScoreID   string      `bson:"score_id"`
	Data      interface{} `bson:"data"`
	IsValid   bool        `bson:"is_valid"`
	Timestamp time.Time   `bson:"timestamp"`
}

type ScoreIDResult struct {
	ScoreID string
	Result  interface{}
	Error   error
}

type Stats struct {
	ProcessedCount  int64
	SuccessCount    int64
	ErrorCount      int64
	InvalidCount    int64
	LastProcessedID string
	StartTime       time.Time
	mu              sync.Mutex
}

func (s *Stats) Update(result ScoreIDResult, isValid bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ProcessedCount++
	s.LastProcessedID = result.ScoreID

	if result.Error != nil {
		s.ErrorCount++
	} else if !isValid {
		s.InvalidCount++
	} else {
		s.SuccessCount++
	}
}

func (s *Stats) GetMetrics() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	duration := time.Since(s.StartTime)
	ratePerSecond := float64(s.ProcessedCount) / duration.Seconds()

	return map[string]interface{}{
		"processed_count":   s.ProcessedCount,
		"success_count":     s.SuccessCount,
		"error_count":       s.ErrorCount,
		"invalid_count":     s.InvalidCount,
		"last_processed_id": s.LastProcessedID,
		"duration_seconds":  duration.Seconds(),
		"rate_per_second":   ratePerSecond,
	}
}

func initLogger() *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{"stdout", "api_cache.log"}

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	return logger.Sugar()
}

func NewAPICacheManager(ctx context.Context, connectionString, dbName string, cacheExpiryHours int) (*APICacheManager, error) {
	logger := initLogger()

	logger.Info("Initializing API Cache Manager",
		"connection_string", connectionString,
		"db_name", dbName,
		"cache_expiry_hours", cacheExpiryHours,
	)

	// Create MongoDB client
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Get collection
	collection := client.Database(dbName).Collection("api_cache")

	// Check existing indexes
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		logger.Error("Failed to list indexes", "error", err)
		return nil, fmt.Errorf("failed to list indexes: %v", err)
	}
	defer cursor.Close(ctx)

	var existingIndexes []bson.M
	if err = cursor.All(ctx, &existingIndexes); err != nil {
		logger.Error("Failed to read indexes", "error", err)
		return nil, fmt.Errorf("failed to read indexes: %v", err)
	}

	// Create map of existing index names
	existingIndexNames := make(map[string]bool)
	for _, idx := range existingIndexes {
		if name, ok := idx["name"].(string); ok {
			existingIndexNames[name] = true
			logger.Debug("Found existing index", "name", name)
		}
	}

	// Create only missing indexes
	if !existingIndexNames["score_id_1"] {
		logger.Info("Creating score_id index")
		_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "score_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			logger.Error("Failed to create score_id index", "error", err)
			return nil, fmt.Errorf("failed to create score_id index: %v", err)
		}
	}

	if !existingIndexNames["timestamp_1"] {
		logger.Info("Creating timestamp index")
		_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "timestamp", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(int32(cacheExpiryHours * 3600)),
		})
		if err != nil {
			logger.Error("Failed to create timestamp index", "error", err)
			return nil, fmt.Errorf("failed to create timestamp index: %v", err)
		}
	}

	// Create HTTP client with timeouts
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &APICacheManager{
		client:          client,
		cacheCollection: collection,
		httpClient:      httpClient,
		rateLimiter:     rate.NewLimiter(rate.Every(time.Second/1000), 1), // 30 requests per second
		logger:          logger,
		cacheExpiry:     time.Duration(cacheExpiryHours) * time.Hour,
	}, nil
}

func (m *APICacheManager) generateNextCombination(current, charset string) string {
	if current == "" {
		return charset[:1]
	}

	chars := []rune(current)
	charsetRunes := []rune(charset)

	for i := len(chars) - 1; i >= 0; i-- {
		pos := strings.IndexRune(charset, chars[i])
		if pos < len(charsetRunes)-1 {
			chars[i] = charsetRunes[pos+1]
			for j := i + 1; j < len(chars); j++ {
				chars[j] = charsetRunes[0]
			}
			return string(chars)
		}
	}

	if len(current) < 32 {
		return strings.Repeat(string(charsetRunes[0]), len(current)+1)
	}
	return ""
}

func (m *APICacheManager) getCachedResult(ctx context.Context, scoreID string) (interface{}, error) {
	m.logger.Debugw("Checking cache", "score_id", scoreID)

	var entry CacheEntry
	err := m.cacheCollection.FindOne(ctx, bson.M{
		"score_id": scoreID,
		"timestamp": bson.M{
			"$gt": time.Now().Add(-m.cacheExpiry),
		},
	}).Decode(&entry)

	if err == mongo.ErrNoDocuments {
		m.logger.Debugw("Cache miss", "score_id", scoreID)
		return nil, nil
	}
	if err != nil {
		m.logger.Warnw("Cache error", "score_id", scoreID, "error", err)
		return nil, err
	}

	m.logger.Debugw("Cache hit", "score_id", scoreID)
	return entry.Data, nil
}

func (m *APICacheManager) cacheResult(ctx context.Context, scoreID string, data interface{}) error {
	m.logger.Debugw("Caching result", "score_id", scoreID)

	entry := CacheEntry{
		ScoreID:   scoreID,
		Data:      data,
		IsValid:   true,
		Timestamp: time.Now(),
	}

	_, err := m.cacheCollection.UpdateOne(ctx,
		bson.M{"score_id": scoreID},
		bson.M{"$set": entry},
		options.Update().SetUpsert(true))

	if err != nil {
		m.logger.Errorw("Failed to cache result", "score_id", scoreID, "error", err)
		return err
	}

	m.logger.Debugw("Result cached successfully", "score_id", scoreID)
	return nil
}

func (m *APICacheManager) fetchAPIResult(ctx context.Context, scoreID string) (*APIResponse, error) {
	// Wait for rate limiter
	err := m.rateLimiter.Wait(ctx)
	if err != nil {
		m.logger.Errorw("Rate limiter error", "score_id", scoreID, "error", err)
		return nil, err
	}

	reqBody := map[string]string{"score_id": scoreID}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		m.logger.Errorw("Failed to marshal request", "score_id", scoreID, "error", err)
		return nil, err
	}

	m.logger.Debugw("Sending API request", "score_id", scoreID)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://yasss.om-api.com/score/advanced/result",
		strings.NewReader(string(reqJSON)))
	if err != nil {
		m.logger.Errorw("Failed to create request", "score_id", scoreID, "error", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.logger.Errorw("API request failed", "score_id", scoreID, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	var result APIResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		m.logger.Errorw("Failed to decode response", "score_id", scoreID, "error", err)
		return nil, err
	}

	m.logger.Debugw("API request successful",
		"score_id", scoreID,
		"status", result.Status,
	)

	return &result, nil
}

func (m *APICacheManager) processScoreID(ctx context.Context, scoreID string) ScoreIDResult {
	m.logger.Debugw("Processing score ID", "score_id", scoreID)

	// Try cache first
	cached, err := m.getCachedResult(ctx, scoreID)
	if err != nil {
		m.logger.Warnw("Cache error", "score_id", scoreID, "error", err)
	} else if cached != nil {
		m.logger.Debugw("Cache hit", "score_id", scoreID)
		return ScoreIDResult{ScoreID: scoreID, Result: cached}
	}

	// Fetch from API
	result, err := m.fetchAPIResult(ctx, scoreID)
	if err != nil {
		m.logger.Errorw("API fetch failed", "score_id", scoreID, "error", err)
		return ScoreIDResult{ScoreID: scoreID, Error: err}
	}

	// Cache the result
	err = m.cacheResult(ctx, scoreID, result)
	if err != nil {
		m.logger.Warnw("Failed to cache result", "score_id", scoreID, "error", err)
	}

	return ScoreIDResult{ScoreID: scoreID, Result: result}
}

func (m *APICacheManager) generatePreviousCombination(current, charset string) string {
	if current == "" {
		return ""
	}

	chars := []rune(current)
	charsetRunes := []rune(charset)

	// Go from right to left
	for i := len(chars) - 1; i >= 0; i-- {
		pos := strings.IndexRune(charset, chars[i])
		if pos > 0 {
			// Go down one character in the charset
			chars[i] = charsetRunes[pos-1]
			// Reset all positions to the right to the highest value
			for j := i + 1; j < len(chars); j++ {
				chars[j] = charsetRunes[len(charsetRunes)-1]
			}
			return string(chars)
		}
	}

	// If we've reached here, we need to reduce the length
	if len(current) > 1 {
		return strings.Repeat(string(charsetRunes[len(charsetRunes)-1]), len(current)-1)
	}
	return ""
}

func (m *APICacheManager) ProcessScoreIDsConcurrent(ctx context.Context, startID string, workers int) (chan ScoreIDResult, *Stats) {
	results := make(chan ScoreIDResult)
	workQueue := make(chan string, workers*2)
	stats := &Stats{StartTime: time.Now()}

	m.logger.Infow("Starting concurrent processing",
		"start_id", startID,
		"workers", workers,
	)

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			m.logger.Infow("Starting worker", "worker_id", workerID)

			for scoreID := range workQueue {
				select {
				case <-ctx.Done():
					m.logger.Infow("Worker stopped", "worker_id", workerID)
					return
				default:
					result := m.processScoreID(ctx, scoreID)
					results <- result

					isValid := false
					if apiResp, ok := result.Result.(*APIResponse); ok {
						isValid = apiResp.Status
					}
					stats.Update(result, isValid)

					if stats.ProcessedCount%100 == 0 {
						metrics := stats.GetMetrics()
						m.logger.Infow("Processing progress",
							"processed", metrics["processed_count"],
							"success", metrics["success_count"],
							"errors", metrics["error_count"],
							"invalid", metrics["invalid_count"],
							"rate", metrics["rate_per_second"],
						)
					}
				}
			}
			m.logger.Infow("Worker finished", "worker_id", workerID)
		}(i)
	}

	// Start ID generator goroutine with descending order
	go func() {
		charset := "abcdefghijklmnopqrstuvwxyz0123456789"
		currentID := startID
		consecutiveInvalid := 0
		maxConsecutiveInvalid := 10000

		for currentID != "" && consecutiveInvalid < maxConsecutiveInvalid {
			select {
			case <-ctx.Done():
				m.logger.Info("ID generator stopped due to context cancellation")
				close(workQueue)
				return
			case workQueue <- currentID:
				currentID = m.generatePreviousCombination(currentID, charset)
				if currentID == "" {
					m.logger.Info("Reached end of ID combinations")
				}
			}
		}

		m.logger.Info("ID generator finished")
		close(workQueue)
	}()

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
		m.logger.Info("All workers finished")

		metrics := stats.GetMetrics()
		m.logger.Infow("Final statistics",
			"total_processed", metrics["processed_count"],
			"total_success", metrics["success_count"],
			"total_errors", metrics["error_count"],
			"total_invalid", metrics["invalid_count"],
			"final_rate", metrics["rate_per_second"],
			"duration_seconds", metrics["duration_seconds"],
		)
	}()

	return results, stats
}

func main() {
	ctx := context.Background()

	manager, err := NewAPICacheManager(ctx, "mongodb://localhost:27017", "yasss_om-api_com", 3000)
	if err != nil {
		manager.logger.Fatalw("Failed to create cache manager", "error", err)
	}

	// Process score IDs concurrently with 10 workers, starting from "biddig" and going down
	resultsChan, stats := manager.ProcessScoreIDsConcurrent(ctx, "biddig", 10)

	// Process results as they come in
	for result := range resultsChan {
		if result.Error != nil {
			manager.logger.Warnw("Error processing score ID",
				"score_id", result.ScoreID,
				"error", result.Error,
			)
			continue
		}

		if apiResp, ok := result.Result.(*APIResponse); ok && apiResp.Status {
			manager.logger.Infow("Valid result found",
				"score_id", result.ScoreID,
				"response", apiResp,
			)
		}
	}

	// Log final statistics
	metrics := stats.GetMetrics()
	manager.logger.Infow("Processing completed",
		"total_processed", metrics["processed_count"],
		"total_success", metrics["success_count"],
		"total_errors", metrics["error_count"],
		"total_invalid", metrics["invalid_count"],
		"final_rate", metrics["rate_per_second"],
		"duration_seconds", metrics["duration_seconds"],
	)
}
