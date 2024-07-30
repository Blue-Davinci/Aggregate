package data

import "math/rand/v2"

// Holds the wight for each unit that we will use in our scoring algorithm.
type weights struct {
	TotalFollows    float64
	TotalLikes      float64
	TotalComments   float64
	EngagementScore float64
	NormalizeScore  float64
	minConsistency  float64
	maxConsistency  float64
	RandomFactor    float64
}

// scoreCalculationAlgorithm() computes the score for a given top creator.
// we compute the engagement score using the formula: (Total Follows * 0.7) + (Total Likes * 0.3)
// we then get the consistency as the ratio of total created feeds to the average time between feeds.
// after which we normalize the consistency score to a 0-100 range.
// and combine the engagement score (80% weight) and the normalized consistency score (20% weight)
// to get the final score.
func (m FeedModel) scoreCalculationAlgorithm(topCreator *TopCreators) float64 {
	weight := weights{
		TotalFollows:    0.5,
		TotalLikes:      0.3,
		TotalComments:   0.2,
		EngagementScore: 0.8,
		NormalizeScore:  0.2,
		minConsistency:  0.0,
		maxConsistency:  100.0,
		RandomFactor:    0.05,
	}
	engagementScore := (float64(topCreator.Total_Follows) * weight.TotalFollows) +
		(float64(topCreator.Total_Likes) * weight.TotalLikes) +
		(float64(topCreator.Total_Comments) * weight.TotalComments)
	// calculate the consistency score
	consistencyScore := 1.0
	if topCreator.Average_Time_Between_Feeds != 0 {
		consistencyScore = float64(topCreator.Total_Created_Feeds) / float64(topCreator.Average_Time_Between_Feeds)
	}
	normalizedConsistency := normalize(consistencyScore, weight.minConsistency, weight.maxConsistency)
	// introduce our random factor and generate a number between 0 and 1
	// we do this to add some randomness to the final score and help differentiate
	// users with similar scores.
	randomFactor := rand.Float64()

	// calculate the final score
	finalScore := (engagementScore * weight.EngagementScore) +
		(normalizedConsistency * weight.NormalizeScore) +
		(randomFactor * weight.RandomFactor)

	return finalScore
}

// normalize() Scales the consistency score to a 0-100 range using Min-Max Normalization.
func normalize(score float64, minScore float64, maxScore float64) float64 {
	// Min-Max Normalization
	if maxScore == minScore {
		return 0 // Avoid division by zero
	}
	normalizedScore := (score - minScore) / (maxScore - minScore) * 100

	// Ensure the normalized score is within 0 to 100 range
	if normalizedScore > 100 {
		return 100
	}
	if normalizedScore < 0 {
		return 0
	}
	return normalizedScore
}
