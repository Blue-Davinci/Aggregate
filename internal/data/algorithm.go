package data

// Holds the wight for each unit that we will use in our scoring algorithm.
type weights struct {
	TotalFollows    float64
	TotalLikes      float64
	EngagementScore float64
	NormalizeScore  float64
}

// scoreCalculationAlgorithm() computes the score for a given top creator.
// we compute the engagement score using the formula: (Total Follows * 0.7) + (Total Likes * 0.3)
// we then get the consistency as the ratio of total created feeds to the average time between feeds.
// after which we normalize the consistency score to a 0-100 range.
// and combine the engagement score (80% weight) and the normalized consistency score (20% weight)
// to get the final score.
func (m FeedModel) scoreCalculationAlgorithm(topCreator *TopCreators) float64 {
	weight := weights{
		TotalFollows:    0.7,
		TotalLikes:      0.3,
		EngagementScore: 0.8,
		NormalizeScore:  0.2,
	}
	engagementScore := (float64(topCreator.Total_Follows) * weight.TotalFollows) + (float64(topCreator.Total_Likes) * weight.TotalLikes)
	consistencyScore := 1.0
	if topCreator.Average_Time_Between_Feeds != 0 {
		consistencyScore = float64(topCreator.Total_Created_Feeds) / float64(topCreator.Average_Time_Between_Feeds)
	}
	normalizedConsistency := normalize(consistencyScore)
	// calculate the final score
	finalScore := (engagementScore * weight.EngagementScore) + (normalizedConsistency * weight.NormalizeScore)
	return finalScore
}

// normalize() Scales the consistency score to a 0-100 range.
// This is a simple normalization logic that can be adjusted based on our requirements.
func normalize(score float64) float64 {
	// Normalization logic here
	// We scale the score to a 0-100 range.
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}
