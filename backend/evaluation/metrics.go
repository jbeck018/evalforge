package evaluation

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

)

// DefaultMetricsCalculator implements the MetricsCalculator interface
type DefaultMetricsCalculator struct{}

// NewMetricsCalculator creates a new metrics calculator
func NewMetricsCalculator() *DefaultMetricsCalculator {
	return &DefaultMetricsCalculator{}
}

// CalculateMetrics calculates comprehensive metrics based on test results
func (mc *DefaultMetricsCalculator) CalculateMetrics(ctx context.Context, testCases []TestCase, analysis *PromptAnalysis) (*EvaluationMetrics, error) {
	if len(testCases) == 0 {
		return nil, fmt.Errorf("no test cases provided for metrics calculation")
	}

	metrics := &EvaluationMetrics{
		ID:             0,
		EvaluationID:   testCases[0].EvaluationID, // Assume all test cases belong to same evaluation
		CustomMetrics:  make(map[string]float64),
		CalculatedAt:   time.Now(),
	}

	// Calculate basic pass/fail metrics
	totalCases := len(testCases)
	passedCases := 0
	totalScore := 0.0
	totalWeight := 0.0

	for _, testCase := range testCases {
		if testCase.Status == "passed" {
			passedCases++
		}
		totalScore += testCase.Score * testCase.Weight
		totalWeight += testCase.Weight
	}

	metrics.TestCasesTotal = totalCases
	metrics.TestCasesPassed = passedCases
	metrics.PassRate = float64(passedCases) / float64(totalCases)
	
	if totalWeight > 0 {
		metrics.OverallScore = totalScore / totalWeight
	}

	// Calculate task-specific metrics
	switch analysis.TaskType {
	case TaskClassification:
		classificationMetrics, err := mc.calculateClassificationMetricsFromTestCases(testCases, analysis)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate classification metrics: %w", err)
		}
		metrics.ClassificationMetrics = classificationMetrics

	case TaskGeneration, TaskSummarization:
		generationMetrics, err := mc.calculateGenerationMetricsFromTestCases(testCases)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate generation metrics: %w", err)
		}
		metrics.GenerationMetrics = generationMetrics

	default:
		// For other task types, calculate custom metrics
		customMetrics, err := mc.calculateTaskSpecificMetrics(testCases, analysis.TaskType)
		if err == nil {
			for key, value := range customMetrics {
				metrics.CustomMetrics[key] = value
			}
		}
	}

	return metrics, nil
}

// CalculateClassificationMetrics calculates classification-specific metrics
func (mc *DefaultMetricsCalculator) CalculateClassificationMetrics(predictions, groundTruth []string, classes []string) (*ClassificationMetrics, error) {
	if len(predictions) != len(groundTruth) {
		return nil, fmt.Errorf("predictions and ground truth must have same length")
	}

	if len(predictions) == 0 {
		return nil, fmt.Errorf("no predictions provided")
	}

	// If classes not provided, infer from data
	if len(classes) == 0 {
		classSet := make(map[string]bool)
		for _, pred := range predictions {
			classSet[pred] = true
		}
		for _, truth := range groundTruth {
			classSet[truth] = true
		}
		
		for class := range classSet {
			classes = append(classes, class)
		}
		sort.Strings(classes)
	}

	metrics := &ClassificationMetrics{
		Precision:       make(map[string]float64),
		Recall:          make(map[string]float64),
		F1Score:         make(map[string]float64),
		ConfusionMatrix: make(map[string]map[string]int),
		Support:         make(map[string]int),
	}

	// Initialize confusion matrix and support
	for _, class := range classes {
		metrics.ConfusionMatrix[class] = make(map[string]int)
		for _, otherClass := range classes {
			metrics.ConfusionMatrix[class][otherClass] = 0
		}
		metrics.Support[class] = 0
	}

	// Build confusion matrix and calculate support
	correct := 0
	for i := range predictions {
		pred := predictions[i]
		truth := groundTruth[i]
		
		// Handle unknown classes by adding them to "other" category
		if metrics.ConfusionMatrix[truth] == nil {
			truth = "other"
		}
		if metrics.ConfusionMatrix[truth][pred] == -1 { // Check if pred exists as key
			pred = "other"
		}
		
		metrics.ConfusionMatrix[truth][pred]++
		metrics.Support[truth]++
		
		if pred == truth {
			correct++
		}
	}

	// Calculate accuracy
	metrics.Accuracy = float64(correct) / float64(len(predictions))

	// Calculate per-class metrics
	var f1Scores []float64
	var weightedF1Sum float64
	totalSupport := 0

	for _, class := range classes {
		tp := metrics.ConfusionMatrix[class][class]
		fp := 0
		fn := 0

		// Calculate false positives and false negatives
		for _, otherClass := range classes {
			if otherClass != class {
				fp += metrics.ConfusionMatrix[otherClass][class] // Predicted as class but actually otherClass
				fn += metrics.ConfusionMatrix[class][otherClass] // Actually class but predicted as otherClass
			}
		}

		// Precision
		if tp+fp > 0 {
			metrics.Precision[class] = float64(tp) / float64(tp+fp)
		} else {
			metrics.Precision[class] = 0.0
		}

		// Recall
		if tp+fn > 0 {
			metrics.Recall[class] = float64(tp) / float64(tp+fn)
		} else {
			metrics.Recall[class] = 0.0
		}

		// F1 Score
		if metrics.Precision[class]+metrics.Recall[class] > 0 {
			metrics.F1Score[class] = 2 * metrics.Precision[class] * metrics.Recall[class] / 
				(metrics.Precision[class] + metrics.Recall[class])
		} else {
			metrics.F1Score[class] = 0.0
		}

		f1Scores = append(f1Scores, metrics.F1Score[class])
		weightedF1Sum += metrics.F1Score[class] * float64(metrics.Support[class])
		totalSupport += metrics.Support[class]
	}

	// Calculate macro F1 (unweighted average)
	if len(f1Scores) > 0 {
		sum := 0.0
		for _, f1 := range f1Scores {
			sum += f1
		}
		metrics.MacroF1 = sum / float64(len(f1Scores))
	}

	// Calculate weighted F1
	if totalSupport > 0 {
		metrics.WeightedF1 = weightedF1Sum / float64(totalSupport)
	}

	return metrics, nil
}

// CalculateGenerationMetrics calculates generation-specific metrics
func (mc *DefaultMetricsCalculator) CalculateGenerationMetrics(predictions, references []string) (*GenerationMetrics, error) {
	if len(predictions) != len(references) {
		return nil, fmt.Errorf("predictions and references must have same length")
	}

	if len(predictions) == 0 {
		return nil, fmt.Errorf("no predictions provided")
	}

	metrics := &GenerationMetrics{}

	// Calculate BLEU score (simplified implementation)
	bleuSum := 0.0
	for i := range predictions {
		bleuSum += mc.calculateSimpleBLEU(predictions[i], references[i])
	}
	metrics.BLEU = bleuSum / float64(len(predictions))

	// Calculate ROUGE scores (simplified implementation)
	rouge1Sum, rouge2Sum, rougeLSum := 0.0, 0.0, 0.0
	for i := range predictions {
		r1, r2, rl := mc.calculateSimpleROUGE(predictions[i], references[i])
		rouge1Sum += r1
		rouge2Sum += r2
		rougeLSum += rl
	}
	metrics.ROUGE1 = rouge1Sum / float64(len(predictions))
	metrics.ROUGE2 = rouge2Sum / float64(len(predictions))
	metrics.ROUGEL = rougeLSum / float64(len(predictions))

	// Calculate lexical diversity
	diversitySum := 0.0
	for _, pred := range predictions {
		diversitySum += mc.calculateLexicalDiversity(pred)
	}
	metrics.Diversity = diversitySum / float64(len(predictions))

	// Calculate coherence (simplified - based on sentence connectivity)
	coherenceSum := 0.0
	for _, pred := range predictions {
		coherenceSum += mc.calculateCoherence(pred)
	}
	metrics.Coherence = coherenceSum / float64(len(predictions))

	// Calculate relevance (based on overlap with reference)
	relevanceSum := 0.0
	for i := range predictions {
		relevanceSum += mc.calculateRelevance(predictions[i], references[i])
	}
	metrics.Relevance = relevanceSum / float64(len(predictions))

	// Note: BERTScore and Perplexity would require additional ML models
	// Setting placeholder values for now
	metrics.BERTScore = metrics.ROUGE1 // Approximation
	metrics.Perplexity = math.Max(1.0, 100.0-metrics.ROUGE1*100) // Inverse approximation

	return metrics, nil
}

// CalculateCustomMetrics calculates custom metrics based on rules
func (mc *DefaultMetricsCalculator) CalculateCustomMetrics(testCases []TestCase, rules map[string]interface{}) (map[string]float64, error) {
	customMetrics := make(map[string]float64)

	// Basic custom metrics
	customMetrics["avg_execution_time"] = mc.calculateAverageExecutionTime(testCases)
	customMetrics["error_rate"] = mc.calculateErrorRate(testCases)
	customMetrics["edge_case_performance"] = mc.calculateCategoryPerformance(testCases, "edge_case")
	customMetrics["adversarial_performance"] = mc.calculateCategoryPerformance(testCases, "adversarial")
	customMetrics["weighted_score"] = mc.calculateWeightedScore(testCases)

	// Apply custom rules if provided
	for ruleName, ruleConfig := range rules {
		if metric, err := mc.applyCustomRule(testCases, ruleName, ruleConfig); err == nil {
			customMetrics[ruleName] = metric
		}
	}

	return customMetrics, nil
}

// Helper methods for classification metrics from test cases
func (mc *DefaultMetricsCalculator) calculateClassificationMetricsFromTestCases(testCases []TestCase, analysis *PromptAnalysis) (*ClassificationMetrics, error) {
	var predictions, groundTruth []string

	for _, testCase := range testCases {
		if testCase.ActualOutput == nil || testCase.ExpectedOutput == nil {
			continue
		}

		// Extract predicted and expected classes
		predClass := mc.extractClassFromOutput(testCase.ActualOutput)
		expectedClass := mc.extractClassFromOutput(testCase.ExpectedOutput)

		if predClass != "" && expectedClass != "" {
			predictions = append(predictions, predClass)
			groundTruth = append(groundTruth, expectedClass)
		}
	}

	if len(predictions) == 0 {
		return nil, fmt.Errorf("no valid classification pairs found in test cases")
	}

	return mc.CalculateClassificationMetrics(predictions, groundTruth, analysis.OutputSchema.Classes)
}

// extractClassFromOutput extracts the class label from output map
func (mc *DefaultMetricsCalculator) extractClassFromOutput(output map[string]interface{}) string {
	// Try common field names for classification output
	fields := []string{"class", "label", "sentiment", "category", "prediction", "result"}

	for _, field := range fields {
		if value, exists := output[field]; exists {
			return fmt.Sprintf("%v", value)
		}
	}

	return ""
}

// Helper methods for generation metrics from test cases
func (mc *DefaultMetricsCalculator) calculateGenerationMetricsFromTestCases(testCases []TestCase) (*GenerationMetrics, error) {
	var predictions, references []string

	for _, testCase := range testCases {
		if testCase.ActualOutput == nil || testCase.ExpectedOutput == nil {
			continue
		}

		pred := mc.extractTextFromOutput(testCase.ActualOutput)
		ref := mc.extractTextFromOutput(testCase.ExpectedOutput)

		if pred != "" && ref != "" {
			predictions = append(predictions, pred)
			references = append(references, ref)
		}
	}

	if len(predictions) == 0 {
		return nil, fmt.Errorf("no valid generation pairs found in test cases")
	}

	return mc.CalculateGenerationMetrics(predictions, references)
}

// extractTextFromOutput extracts text content from output map
func (mc *DefaultMetricsCalculator) extractTextFromOutput(output map[string]interface{}) string {
	// Try common field names for text output
	fields := []string{"text", "result", "output", "response", "summary", "answer", "generated"}

	for _, field := range fields {
		if value, exists := output[field]; exists {
			return fmt.Sprintf("%v", value)
		}
	}

	return ""
}

// Task-specific metrics calculation
func (mc *DefaultMetricsCalculator) calculateTaskSpecificMetrics(testCases []TestCase, taskType TaskType) (map[string]float64, error) {
	metrics := make(map[string]float64)

	switch taskType {
	case TaskExtraction:
		metrics["extraction_precision"] = mc.calculateExtractionPrecision(testCases)
		metrics["extraction_recall"] = mc.calculateExtractionRecall(testCases)
		metrics["extraction_f1"] = mc.calculateExtractionF1(testCases)

	case TaskQA:
		metrics["answer_accuracy"] = mc.calculateAnswerAccuracy(testCases)
		metrics["answer_completeness"] = mc.calculateAnswerCompleteness(testCases)

	case TaskTransformation:
		metrics["format_compliance"] = mc.calculateFormatCompliance(testCases)
		metrics["content_preservation"] = mc.calculateContentPreservation(testCases)
	}

	return metrics, nil
}

// Simplified BLEU calculation
func (mc *DefaultMetricsCalculator) calculateSimpleBLEU(prediction, reference string) float64 {
	predTokens := strings.Fields(strings.ToLower(prediction))
	refTokens := strings.Fields(strings.ToLower(reference))

	if len(predTokens) == 0 || len(refTokens) == 0 {
		return 0.0
	}

	// Simple unigram precision
	refTokenSet := make(map[string]int)
	for _, token := range refTokens {
		refTokenSet[token]++
	}

	matches := 0
	for _, token := range predTokens {
		if refTokenSet[token] > 0 {
			matches++
			refTokenSet[token]--
		}
	}

	precision := float64(matches) / float64(len(predTokens))
	
	// Apply brevity penalty
	bp := 1.0
	if len(predTokens) < len(refTokens) {
		bp = math.Exp(1.0 - float64(len(refTokens))/float64(len(predTokens)))
	}

	return bp * precision
}

// Simplified ROUGE calculation
func (mc *DefaultMetricsCalculator) calculateSimpleROUGE(prediction, reference string) (rouge1, rouge2, rougeL float64) {
	predTokens := strings.Fields(strings.ToLower(prediction))
	refTokens := strings.Fields(strings.ToLower(reference))

	// ROUGE-1 (unigram overlap)
	rouge1 = mc.calculateTokenOverlap(predTokens, refTokens)

	// ROUGE-2 (bigram overlap) - simplified
	if len(predTokens) > 1 && len(refTokens) > 1 {
		predBigrams := mc.getBigrams(predTokens)
		refBigrams := mc.getBigrams(refTokens)
		rouge2 = mc.calculateTokenOverlap(predBigrams, refBigrams)
	}

	// ROUGE-L (longest common subsequence) - approximated
	rougeL = mc.calculateLCS(predTokens, refTokens)

	return rouge1, rouge2, rougeL
}

// Helper methods for ROUGE calculation
func (mc *DefaultMetricsCalculator) calculateTokenOverlap(tokens1, tokens2 []string) float64 {
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0.0
	}

	set1 := make(map[string]bool)
	for _, token := range tokens1 {
		set1[token] = true
	}

	overlap := 0
	for _, token := range tokens2 {
		if set1[token] {
			overlap++
		}
	}

	return float64(overlap) / float64(len(tokens2))
}

func (mc *DefaultMetricsCalculator) getBigrams(tokens []string) []string {
	if len(tokens) < 2 {
		return []string{}
	}

	bigrams := make([]string, len(tokens)-1)
	for i := 0; i < len(tokens)-1; i++ {
		bigrams[i] = tokens[i] + " " + tokens[i+1]
	}
	return bigrams
}

func (mc *DefaultMetricsCalculator) calculateLCS(tokens1, tokens2 []string) float64 {
	// Simplified LCS calculation
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0.0
	}

	lcsLength := mc.longestCommonSubsequence(tokens1, tokens2)
	return float64(lcsLength) / float64(max(len(tokens1), len(tokens2)))
}

func (mc *DefaultMetricsCalculator) longestCommonSubsequence(a, b []string) int {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}

// Additional metric calculations
func (mc *DefaultMetricsCalculator) calculateLexicalDiversity(text string) float64 {
	tokens := strings.Fields(strings.ToLower(text))
	if len(tokens) == 0 {
		return 0.0
	}

	uniqueTokens := make(map[string]bool)
	for _, token := range tokens {
		uniqueTokens[token] = true
	}

	return float64(len(uniqueTokens)) / float64(len(tokens))
}

func (mc *DefaultMetricsCalculator) calculateCoherence(text string) float64 {
	sentences := strings.Split(text, ".")
	if len(sentences) < 2 {
		return 1.0 // Single sentence is perfectly coherent
	}

	// Simplified coherence based on sentence length variation
	lengths := make([]int, 0, len(sentences))
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if trimmed != "" {
			lengths = append(lengths, len(strings.Fields(trimmed)))
		}
	}

	if len(lengths) < 2 {
		return 1.0
	}

	// Calculate coefficient of variation (lower is more coherent)
	mean := 0.0
	for _, length := range lengths {
		mean += float64(length)
	}
	mean /= float64(len(lengths))

	variance := 0.0
	for _, length := range lengths {
		variance += math.Pow(float64(length)-mean, 2)
	}
	variance /= float64(len(lengths))

	cv := math.Sqrt(variance) / mean
	return math.Max(0.0, 1.0-cv/2.0) // Normalize to 0-1 range
}

func (mc *DefaultMetricsCalculator) calculateRelevance(prediction, reference string) float64 {
	return mc.calculateTokenOverlap(
		strings.Fields(strings.ToLower(prediction)),
		strings.Fields(strings.ToLower(reference)),
	)
}

// Custom metrics helpers
func (mc *DefaultMetricsCalculator) calculateAverageExecutionTime(testCases []TestCase) float64 {
	if len(testCases) == 0 {
		return 0.0
	}

	totalTime := 0.0
	validCases := 0

	for _, testCase := range testCases {
		if testCase.ExecutedAt != nil {
			// This would require storing execution start time as well
			// For now, return a placeholder
			totalTime += 1.0 // Placeholder
			validCases++
		}
	}

	if validCases == 0 {
		return 0.0
	}

	return totalTime / float64(validCases)
}

func (mc *DefaultMetricsCalculator) calculateErrorRate(testCases []TestCase) float64 {
	if len(testCases) == 0 {
		return 0.0
	}

	errorCount := 0
	for _, testCase := range testCases {
		if testCase.Status == "error" || testCase.Status == "failed" {
			errorCount++
		}
	}

	return float64(errorCount) / float64(len(testCases))
}

func (mc *DefaultMetricsCalculator) calculateCategoryPerformance(testCases []TestCase, category string) float64 {
	categoryCases := 0
	categoryPassed := 0

	for _, testCase := range testCases {
		if testCase.Category == category {
			categoryCases++
			if testCase.Status == "passed" {
				categoryPassed++
			}
		}
	}

	if categoryCases == 0 {
		return 0.0
	}

	return float64(categoryPassed) / float64(categoryCases)
}

func (mc *DefaultMetricsCalculator) calculateWeightedScore(testCases []TestCase) float64 {
	totalScore := 0.0
	totalWeight := 0.0

	for _, testCase := range testCases {
		totalScore += testCase.Score * testCase.Weight
		totalWeight += testCase.Weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

func (mc *DefaultMetricsCalculator) applyCustomRule(testCases []TestCase, ruleName string, ruleConfig interface{}) (float64, error) {
	// Placeholder for custom rule application
	// This would be expanded based on specific rule types
	return 0.0, fmt.Errorf("custom rule not implemented: %s", ruleName)
}

// Task-specific metric implementations
func (mc *DefaultMetricsCalculator) calculateExtractionPrecision(testCases []TestCase) float64 {
	// Placeholder implementation
	return mc.calculateCategoryPerformance(testCases, "normal")
}

func (mc *DefaultMetricsCalculator) calculateExtractionRecall(testCases []TestCase) float64 {
	// Placeholder implementation
	return mc.calculateCategoryPerformance(testCases, "edge_case")
}

func (mc *DefaultMetricsCalculator) calculateExtractionF1(testCases []TestCase) float64 {
	precision := mc.calculateExtractionPrecision(testCases)
	recall := mc.calculateExtractionRecall(testCases)
	
	if precision+recall == 0 {
		return 0.0
	}
	
	return 2 * precision * recall / (precision + recall)
}

func (mc *DefaultMetricsCalculator) calculateAnswerAccuracy(testCases []TestCase) float64 {
	return mc.calculateCategoryPerformance(testCases, "normal")
}

func (mc *DefaultMetricsCalculator) calculateAnswerCompleteness(testCases []TestCase) float64 {
	// Placeholder - would measure how complete answers are
	return 0.8
}

func (mc *DefaultMetricsCalculator) calculateFormatCompliance(testCases []TestCase) float64 {
	// Placeholder - would check if output follows expected format
	return 0.9
}

func (mc *DefaultMetricsCalculator) calculateContentPreservation(testCases []TestCase) float64 {
	// Placeholder - would measure how well content is preserved in transformation
	return 0.85
}

// Utility functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}