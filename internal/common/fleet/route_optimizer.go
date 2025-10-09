package fleet

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// RouteOptimizer provides route optimization capabilities
type RouteOptimizer struct {
	db    *gorm.DB
	redis *redis.Client
}

// RouteRequest represents a route optimization request
type RouteRequest struct {
	CompanyID     string                 `json:"company_id"`
	VehicleID     string                 `json:"vehicle_id"`
	DriverID      string                 `json:"driver_id"`
	Stops         []RouteStop            `json:"stops"`
	Constraints   RouteConstraints       `json:"constraints"`
	Optimization  OptimizationCriteria   `json:"optimization"`
	TimeWindow    *TimeWindow            `json:"time_window,omitempty"`
}

// RouteStop represents a stop in the route
type RouteStop struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Priority    int       `json:"priority"`    // 1-10, higher is more important
	ServiceTime int       `json:"service_time"` // minutes
	TimeWindow  TimeWindow `json:"time_window,omitempty"`
	Type        string    `json:"type"` // pickup, delivery, service, fuel
}

// RouteConstraints represents constraints for route optimization
type RouteConstraints struct {
	MaxDistance     float64 `json:"max_distance"`     // km
	MaxDuration     int     `json:"max_duration"`     // minutes
	MaxStops        int     `json:"max_stops"`
	VehicleCapacity float64 `json:"vehicle_capacity"` // kg or m³
	DriverHours     int     `json:"driver_hours"`     // max working hours
	FuelLimit       float64 `json:"fuel_limit"`       // liters
}

// OptimizationCriteria defines what to optimize for
type OptimizationCriteria struct {
	MinimizeDistance bool `json:"minimize_distance"`
	MinimizeTime     bool `json:"minimize_time"`
	MinimizeFuel     bool `json:"minimize_fuel"`
	MaximizeEfficiency bool `json:"maximize_efficiency"`
	BalanceLoad      bool `json:"balance_load"`
}

// TimeWindow represents a time window constraint
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// OptimizedRoute represents the result of route optimization
type OptimizedRoute struct {
	ID              string      `json:"id"`
	VehicleID       string      `json:"vehicle_id"`
	DriverID        string      `json:"driver_id"`
	Stops           []RouteStop `json:"stops"`
	TotalDistance   float64     `json:"total_distance"`   // km
	TotalDuration   int         `json:"total_duration"`   // minutes
	TotalFuelCost   float64     `json:"total_fuel_cost"`  // IDR
	EstimatedArrival time.Time  `json:"estimated_arrival"`
	OptimizationScore float64   `json:"optimization_score"`
	CreatedAt       time.Time   `json:"created_at"`
}

// RouteNode represents a node in the route graph
type RouteNode struct {
	Stop      RouteStop
	Distance  float64
	Duration  int
	FuelCost  float64
	Visited   bool
}

// NewRouteOptimizer creates a new route optimizer
func NewRouteOptimizer(db *gorm.DB, redis *redis.Client) *RouteOptimizer {
	return &RouteOptimizer{
		db:    db,
		redis: redis,
	}
}

// OptimizeRoute optimizes a route based on the given criteria
func (ro *RouteOptimizer) OptimizeRoute(ctx context.Context, req *RouteRequest) (*OptimizedRoute, error) {
	if len(req.Stops) < 2 {
		return nil, fmt.Errorf("at least 2 stops required for route optimization")
	}

	// Validate constraints
	if err := ro.validateConstraints(req); err != nil {
		return nil, fmt.Errorf("constraint validation failed: %w", err)
	}

	// Check cache first
	cacheKey := ro.generateCacheKey(req)
	cachedRoute, err := ro.getCachedRoute(ctx, cacheKey)
	if err == nil && cachedRoute != nil {
		return cachedRoute, nil
	}

	// Generate optimized route
	optimizedRoute, err := ro.generateOptimizedRoute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate optimized route: %w", err)
	}

	// Cache the result
	ro.cacheRoute(ctx, cacheKey, optimizedRoute, 1*time.Hour)

	return optimizedRoute, nil
}

// validateConstraints validates route constraints
func (ro *RouteOptimizer) validateConstraints(req *RouteRequest) error {
	if req.Constraints.MaxDistance <= 0 {
		return fmt.Errorf("max distance must be positive")
	}
	if req.Constraints.MaxDuration <= 0 {
		return fmt.Errorf("max duration must be positive")
	}
	if req.Constraints.MaxStops <= 0 {
		return fmt.Errorf("max stops must be positive")
	}
	if len(req.Stops) > req.Constraints.MaxStops {
		return fmt.Errorf("number of stops (%d) exceeds max stops (%d)", len(req.Stops), req.Constraints.MaxStops)
	}
	return nil
}

// generateOptimizedRoute generates an optimized route using multiple algorithms
func (ro *RouteOptimizer) generateOptimizedRoute(_ context.Context, req *RouteRequest) (*OptimizedRoute, error) {
	// Try different optimization algorithms and pick the best result
	var bestRoute *OptimizedRoute
	var bestScore float64 = -1

	// Algorithm 1: Nearest Neighbor (fast, good for small routes)
	if len(req.Stops) <= 10 {
		route, err := ro.nearestNeighborOptimization(req)
		if err == nil {
			score := ro.calculateOptimizationScore(route, req.Optimization)
			if score > bestScore {
				bestRoute = route
				bestScore = score
			}
		}
	}

	// Algorithm 2: Genetic Algorithm (better for complex routes)
	if len(req.Stops) > 5 {
		route, err := ro.geneticAlgorithmOptimization(req)
		if err == nil {
			score := ro.calculateOptimizationScore(route, req.Optimization)
			if score > bestScore {
				bestRoute = route
				bestScore = score
			}
		}
	}

	// Algorithm 3: Simulated Annealing (good balance)
	route, err := ro.simulatedAnnealingOptimization(req)
	if err == nil {
		score := ro.calculateOptimizationScore(route, req.Optimization)
		if score > bestScore {
			bestRoute = route
			bestScore = score
		}
	}

	if bestRoute == nil {
		return nil, fmt.Errorf("failed to generate any valid route")
	}

	bestRoute.OptimizationScore = bestScore
	bestRoute.ID = fmt.Sprintf("route_%d", time.Now().UnixNano())
	bestRoute.CreatedAt = time.Now()

	return bestRoute, nil
}

// nearestNeighborOptimization implements nearest neighbor algorithm
func (ro *RouteOptimizer) nearestNeighborOptimization(req *RouteRequest) (*OptimizedRoute, error) {
	if len(req.Stops) < 2 {
		return nil, fmt.Errorf("insufficient stops for optimization")
	}

	// Start with the first stop
	route := &OptimizedRoute{
		VehicleID: req.VehicleID,
		DriverID:  req.DriverID,
		Stops:     make([]RouteStop, 0, len(req.Stops)),
	}

	visited := make(map[string]bool)
	currentStop := req.Stops[0]
	route.Stops = append(route.Stops, currentStop)
	visited[currentStop.ID] = true

	// Find nearest unvisited stop iteratively
	for len(visited) < len(req.Stops) {
		nearestStop := ro.findNearestStop(currentStop, req.Stops, visited)
		if nearestStop == nil {
			break
		}

		route.Stops = append(route.Stops, *nearestStop)
		visited[nearestStop.ID] = true
		currentStop = *nearestStop
	}

	// Calculate route metrics
	ro.calculateRouteMetrics(route)

	return route, nil
}

// geneticAlgorithmOptimization implements genetic algorithm for route optimization
func (ro *RouteOptimizer) geneticAlgorithmOptimization(req *RouteRequest) (*OptimizedRoute, error) {
	const (
		populationSize = 50
		generations    = 100
		mutationRate   = 0.1
		crossoverRate  = 0.8
	)

	// Initialize population
	population := ro.initializePopulation(req.Stops, populationSize)

	// Evolve population
	for generation := 0; generation < generations; generation++ {
		// Evaluate fitness
		for i := range population {
			population[i].Fitness = ro.calculateFitness(population[i], req.Optimization)
		}

		// Sort by fitness (higher is better)
		sort.Slice(population, func(i, j int) bool {
			return population[i].Fitness > population[j].Fitness
		})

		// Create new generation
		newPopulation := make([]RouteChromosome, populationSize)
		
		// Keep best 10% (elitism)
		eliteCount := populationSize / 10
		copy(newPopulation[:eliteCount], population[:eliteCount])

		// Generate offspring
		for i := eliteCount; i < populationSize; i++ {
			parent1 := ro.tournamentSelection(population)
			parent2 := ro.tournamentSelection(population)
			
			if ro.randomFloat() < crossoverRate {
				child := ro.crossover(parent1, parent2)
				if ro.randomFloat() < mutationRate {
					child = ro.mutate(child)
				}
				newPopulation[i] = child
			} else {
				newPopulation[i] = parent1
			}
		}

		population = newPopulation
	}

	// Return best route
	bestChromosome := population[0]
	route := &OptimizedRoute{
		VehicleID: req.VehicleID,
		DriverID:  req.DriverID,
		Stops:     bestChromosome.Stops,
	}

	ro.calculateRouteMetrics(route)
	return route, nil
}

// simulatedAnnealingOptimization implements simulated annealing algorithm
func (ro *RouteOptimizer) simulatedAnnealingOptimization(req *RouteRequest) (*OptimizedRoute, error) {
	const (
		initialTemp = 1000.0
		finalTemp   = 0.1
		coolingRate = 0.95
	)

	// Start with a random route
	currentRoute := ro.generateRandomRoute(req.Stops)
	currentCost := ro.calculateRouteCost(currentRoute, req.Optimization)

	bestRoute := currentRoute
	bestCost := currentCost

	temperature := initialTemp

	for temperature > finalTemp {
		// Generate neighbor solution
		neighborRoute := ro.generateNeighbor(currentRoute)
		neighborCost := ro.calculateRouteCost(neighborRoute, req.Optimization)

		// Accept or reject neighbor
		if neighborCost < currentCost || ro.randomFloat() < math.Exp(-(neighborCost-currentCost)/temperature) {
			currentRoute = neighborRoute
			currentCost = neighborCost

			if currentCost < bestCost {
				bestRoute = currentRoute
				bestCost = currentCost
			}
		}

		temperature *= coolingRate
	}

	route := &OptimizedRoute{
		VehicleID: req.VehicleID,
		DriverID:  req.DriverID,
		Stops:     bestRoute,
	}

	ro.calculateRouteMetrics(route)
	return route, nil
}

// RouteChromosome represents a route in genetic algorithm
type RouteChromosome struct {
	Stops   []RouteStop `json:"stops"`
	Fitness float64     `json:"fitness"`
}

// initializePopulation creates initial population for genetic algorithm
func (ro *RouteOptimizer) initializePopulation(stops []RouteStop, size int) []RouteChromosome {
	population := make([]RouteChromosome, size)
	
	for i := 0; i < size; i++ {
		population[i] = RouteChromosome{
			Stops: ro.generateRandomRoute(stops),
		}
	}
	
	return population
}

// generateRandomRoute creates a random route from stops
func (ro *RouteOptimizer) generateRandomRoute(stops []RouteStop) []RouteStop {
	route := make([]RouteStop, len(stops))
	copy(route, stops)
	
	// Shuffle the route
	for i := len(route) - 1; i > 0; i-- {
		j := ro.randomInt(i + 1)
		route[i], route[j] = route[j], route[i]
	}
	
	return route
}

// calculateFitness calculates fitness of a route chromosome
func (ro *RouteOptimizer) calculateFitness(chromosome RouteChromosome, criteria OptimizationCriteria) float64 {
	// Calculate route cost (lower is better)
	cost := ro.calculateRouteCost(chromosome.Stops, criteria)
	
	// Convert to fitness (higher is better)
	return 1.0 / (1.0 + cost)
}

// calculateRouteCost calculates the cost of a route
func (ro *RouteOptimizer) calculateRouteCost(stops []RouteStop, criteria OptimizationCriteria) float64 {
	if len(stops) < 2 {
		return 0
	}

	var totalCost float64

	// Calculate distance cost
	if criteria.MinimizeDistance {
		distance := ro.calculateTotalDistance(stops)
		totalCost += distance * 0.1 // Weight factor
	}

	// Calculate time cost
	if criteria.MinimizeTime {
		duration := ro.calculateTotalDuration(stops)
		totalCost += float64(duration) * 0.01 // Weight factor
	}

	// Calculate fuel cost
	if criteria.MinimizeFuel {
		fuelCost := ro.calculateTotalFuelCost(stops)
		totalCost += fuelCost * 0.001 // Weight factor
	}

	return totalCost
}

// calculateOptimizationScore calculates optimization score for a route
func (ro *RouteOptimizer) calculateOptimizationScore(route *OptimizedRoute, criteria OptimizationCriteria) float64 {
	score := 0.0

	if criteria.MinimizeDistance {
		// Lower distance = higher score
		score += 100.0 / (1.0 + route.TotalDistance)
	}

	if criteria.MinimizeTime {
		// Lower duration = higher score
		score += 100.0 / (1.0 + float64(route.TotalDuration)/60.0)
	}

	if criteria.MinimizeFuel {
		// Lower fuel cost = higher score
		score += 100.0 / (1.0 + route.TotalFuelCost/1000.0)
	}

	if criteria.MaximizeEfficiency {
		// Higher efficiency = higher score
		efficiency := route.TotalDistance / float64(route.TotalDuration) * 60 // km/h
		score += efficiency * 0.1
	}

	return score
}

// Helper methods for route calculations
func (ro *RouteOptimizer) findNearestStop(current RouteStop, stops []RouteStop, visited map[string]bool) *RouteStop {
	var nearest *RouteStop
	minDistance := math.MaxFloat64

	for _, stop := range stops {
		if visited[stop.ID] {
			continue
		}

		distance := ro.calculateDistance(current.Latitude, current.Longitude, stop.Latitude, stop.Longitude)
		if distance < minDistance {
			minDistance = distance
			nearest = &stop
		}
	}

	return nearest
}

func (ro *RouteOptimizer) calculateTotalDistance(stops []RouteStop) float64 {
	if len(stops) < 2 {
		return 0
	}

	totalDistance := 0.0
	for i := 1; i < len(stops); i++ {
		distance := ro.calculateDistance(
			stops[i-1].Latitude, stops[i-1].Longitude,
			stops[i].Latitude, stops[i].Longitude,
		)
		totalDistance += distance
	}

	return totalDistance
}

func (ro *RouteOptimizer) calculateTotalDuration(stops []RouteStop) int {
	if len(stops) < 2 {
		return 0
	}

	totalDuration := 0
	for i := 1; i < len(stops); i++ {
		distance := ro.calculateDistance(
			stops[i-1].Latitude, stops[i-1].Longitude,
			stops[i].Latitude, stops[i].Longitude,
		)
		// Assume average speed of 40 km/h in city traffic
		duration := int(distance / 40.0 * 60) // minutes
		totalDuration += duration
		totalDuration += stops[i].ServiceTime
	}

	return totalDuration
}

func (ro *RouteOptimizer) calculateTotalFuelCost(stops []RouteStop) float64 {
	distance := ro.calculateTotalDistance(stops)
	// Assume 10 km/liter fuel efficiency and 15,000 IDR/liter
	fuelConsumption := distance / 10.0
	fuelCost := fuelConsumption * 15000
	return fuelCost
}

func (ro *RouteOptimizer) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (ro *RouteOptimizer) calculateRouteMetrics(route *OptimizedRoute) {
	route.TotalDistance = ro.calculateTotalDistance(route.Stops)
	route.TotalDuration = ro.calculateTotalDuration(route.Stops)
	route.TotalFuelCost = ro.calculateTotalFuelCost(route.Stops)
	
	// Calculate estimated arrival time
	route.EstimatedArrival = time.Now().Add(time.Duration(route.TotalDuration) * time.Minute)
}

// Genetic algorithm helper methods
func (ro *RouteOptimizer) tournamentSelection(population []RouteChromosome) RouteChromosome {
	const tournamentSize = 3
	
	best := population[ro.randomInt(len(population))]
	for i := 1; i < tournamentSize; i++ {
		candidate := population[ro.randomInt(len(population))]
		if candidate.Fitness > best.Fitness {
			best = candidate
		}
	}
	
	return best
}

func (ro *RouteOptimizer) crossover(parent1, parent2 RouteChromosome) RouteChromosome {
	// Order crossover (OX)
	if len(parent1.Stops) < 2 {
		return parent1
	}

	start := ro.randomInt(len(parent1.Stops))
	end := start + ro.randomInt(len(parent1.Stops)-start)

	child := RouteChromosome{
		Stops: make([]RouteStop, len(parent1.Stops)),
	}

	// Copy segment from parent1
	copy(child.Stops[start:end], parent1.Stops[start:end])

	// Fill remaining positions from parent2
	childIndex := end
	for _, stop := range parent2.Stops {
		if !ro.containsStop(child.Stops[start:end], stop) {
			if childIndex >= len(child.Stops) {
				childIndex = 0
			}
			child.Stops[childIndex] = stop
			childIndex++
		}
	}

	return child
}

func (ro *RouteOptimizer) mutate(chromosome RouteChromosome) RouteChromosome {
	if len(chromosome.Stops) < 2 {
		return chromosome
	}

	// Swap mutation
	i := ro.randomInt(len(chromosome.Stops))
	j := ro.randomInt(len(chromosome.Stops))
	
	chromosome.Stops[i], chromosome.Stops[j] = chromosome.Stops[j], chromosome.Stops[i]
	
	return chromosome
}

func (ro *RouteOptimizer) generateNeighbor(route []RouteStop) []RouteStop {
	if len(route) < 2 {
		return route
	}

	neighbor := make([]RouteStop, len(route))
	copy(neighbor, route)

	// 2-opt swap
	i := ro.randomInt(len(neighbor) - 1)
	j := i + 1 + ro.randomInt(len(neighbor)-i-1)

	// Reverse the segment between i and j
	for k := 0; k < (j-i+1)/2; k++ {
		neighbor[i+k], neighbor[j-k] = neighbor[j-k], neighbor[i+k]
	}

	return neighbor
}

// Utility methods
func (ro *RouteOptimizer) containsStop(stops []RouteStop, stop RouteStop) bool {
	for _, s := range stops {
		if s.ID == stop.ID {
			return true
		}
	}
	return false
}

func (ro *RouteOptimizer) randomFloat() float64 {
	return rand.Float64()
}

func (ro *RouteOptimizer) randomInt(max int) int {
	return rand.Intn(max)
}

// Cache methods
func (ro *RouteOptimizer) generateCacheKey(req *RouteRequest) string {
	// Create a hash of the request parameters for caching
	return fmt.Sprintf("route_opt:%s:%s:%d", req.CompanyID, req.VehicleID, len(req.Stops))
}

func (ro *RouteOptimizer) getCachedRoute(_ context.Context, _ string) (*OptimizedRoute, error) {
	// Implementation would use Redis to get cached route
	// For now, return nil to indicate cache miss
	return nil, fmt.Errorf("cache miss")
}

func (ro *RouteOptimizer) cacheRoute(_ context.Context, _ string, _ *OptimizedRoute, _ time.Duration) error {
	// Implementation would use Redis to cache the route
	// For now, just return nil
	return nil
}
