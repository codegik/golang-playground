package main

// PowerStation
/**
The minimum powered city is the city with the lowest power value across all cities.
Breaking it down:
1. Each city has a power value = sum of all stations within range r
2. The minimum powered city = the city with the smallest power value
3. Goal: Make this minimum as high as possible by adding k stations optimally

Array index is the city number.
Array value is the power of the city.
**/

// PowerStation represents a power station addition
type PowerStation struct {
	pos int // position where station was added
	val int // amount of power added
}

func maxPower(stations []int, rang int, extraStations int) int {
	stationLen := len(stations)
	prefix := make([]int, stationLen+1)
	power := make([]int, stationLen)

	// Constructs a cumulative sum array to quickly calculate the power for any city
	// If stations = [1, 2, 3, 4], then prefix = [0, 1, 3, 6, 10]
	for i := 0; i < stationLen; i++ {
		prefix[i+1] = prefix[i] + stations[i]
	}

	// For each city i, calculates the total power it receives from stations within range
	// The prefix sum is needed for efficient calculation of the initial power at each city.
	// Without it, you'd need nested loops, making the algorithm slower.
	for i := 0; i < stationLen; i++ {
		left := max(0, i-rang)
		right := min(stationLen-1, i+rang)
		power[i] = prefix[right+1] - prefix[left]
	}

	// Find the maximum minimum power that can be achieved with extra stations
	canAchieve := func(minPower int) bool {
		// Queue to track additions of power stations and their effects on cities.
		// The queue acts as a cache of recent additions that might still affect nearby cities,
		// avoiding the need to scan all previous additions.
		queue := NewQueue[PowerStation]()
		// The currentPower copy is needed because canAchieve is called multiple times during binary search,
		// and each call must start with the original power distribution.
		currentPower := make([]int, stationLen)
		copy(currentPower, power)
		// variable tracks the total number of extra stations consumed so far during the simulation
		// to ensure you don't exceed the extraStations budget
		usedStations := 0

		for i := 0; i < stationLen; i++ {
			// Calculate current power including recent additions from queue
			queueSum := 0
			queue.ForEach(func(powerStation PowerStation) {
				// Only count if powerStation affects the current city
				if powerStation.pos >= i-rang && powerStation.pos <= i+rang {
					queueSum += powerStation.val
				}
			})

			totalPower := currentPower[i] + queueSum

			// If a city doesn't reach the target, add stations
			if totalPower < minPower {
				need := minPower - totalPower
				usedStations += need
				if usedStations > extraStations {
					return false // You've tried to place more stations than you have available
				}

				// Add a station at the rightmost position within range covering both cases: Not at edge and At the Edge
				rightmost := min(stationLen-1, i+rang)
				queue.Enqueue(PowerStation{pos: rightmost, val: need})

				// Update power for all cities affected by this addition
				// Example: If rightmost = 5, rang = 2, and we add need = 3 stations:
				// Cities 3, 4, 5, 6, 7 (if they exist) all benefit because they're within range 2 of position 5
				// Each gets +3 power added to their currentPower
				// Why it's needed: When you place stations at position rightmost, all cities within rang distance can use those stations.
				// This loop ensures their power values are immediately updated so future iterations see the correct power levels.
				for j := max(0, rightmost-rang); j <= min(stationLen-1, rightmost+rang); j++ {
					currentPower[j] += need
				}
			}
		}

		return true
	}

	// Binary search for the maximum achievable minimum power
	// The right value represents the theoretical maximum possible power, not just what's in the power array.
	left := 0
	right := 0
	for _, p := range power {
		right += p
	}
	right += extraStations

	result := 0
	for left <= right {
		mid := left + (right-left)/2
		if canAchieve(mid) {
			result = mid
			left = mid + 1 // Try a higher target
		} else {
			right = mid - 1 // Try a lower target
		}
	}

	return result
}
