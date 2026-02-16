# Maximize the Minimum Powered City

You are given a **0-indexed** integer array stations of length `n`, where `stations[i]` represents the number of power stations in the `i` city.

Each power station can provide power to every city in a fixed range. In other words, if the range is denoted by `r`, then a power station at city `i` can provide power to all cities `j` such that `|i - j| <= r` and `0 <= i, j <= n - 1`.

Note that `|x|` denotes absolute value. For example, `|7 - 5| = 2` and `|3 - 10| = 7`.

The power of a city is the total number of power stations it is being provided power from.

The government has sanctioned building `k` more power stations, each of which can be built in any city, and have the same range as the pre-existing ones.

Given the two integers `r` and `k`, return the maximum possible minimum power of a city, if the additional power stations are built optimally.

Note that you can build the `k` power stations in multiple cities.

## Example 1
```
Input: stations = [1,2,4,5,0], r = 1, k = 2
Output: 5
Explanation: 
One of the optimal ways is to install both the power stations at city 1. 
So stations will become [1,4,4,5,0].
- City 0 is provided by 1 + 4 = 5 power stations.
- City 1 is provided by 1 + 4 + 4 = 9 power stations.
- City 2 is provided by 4 + 4 + 5 = 13 power stations.
- City 3 is provided by 5 + 4 = 9 power stations.
- City 4 is provided by 5 + 0 = 5 power stations.
So the minimum power of a city is 5.
Since it is not possible to obtain a larger power, we return 5.
```

## Example 2
```
Input: stations = [4,4,4,4], r = 0, k = 3
Output: 4
Explanation: 
It can be proved that we cannot make the minimum power of a city greater than 4.
```

## Approach

To solve the problem we can use the following approach:
- Binary search: Try different minimum power values to find the maximum achievable.
- Queue: Track station additions efficiently as you simulate checking if a target minimum is achievable

### Power Coverage (Range r)
- A station at city `i` powers all cities within distance `r`
- Distance is calculated as `|i - j|` (absolute difference)
- Example: `If r = 1`, a station at city `2` powers cities `1, 2, and 3`

### City Power Calculation
- A city's power = total stations that can reach it
- Example with stations = [1,2,4,5,0] and r = 1:
- City 0 power = stations[0] + stations[1] = 1 + 2 = 3
- City 1 power = stations[0] + stations[1] + stations[2] = 1 + 2 + 4 = 7
- City 2 power = stations[1] + stations[2] + stations[3] = 2 + 4 + 5 = 11

### Goal
- Build `k` new stations anywhere
- Maximize the minimum power (boost the weakest city as much as possible)

Given 
```
stations = [1,2,4,5,0], r = 1, k = 2
```
The initial power of each city is:
```
City 0: 1 + 2 = 3
City 1: 1 + 2 + 4 = 7
City 2: 2 + 4 + 5 = 11
City 3: 4 + 5 + 0 = 9
City 4: 5 + 0 = 5
```

Then
```
Minimum = 3 (city 0 is weakest)
```

After adding 2 stations at city 1:
```
stations = [1,4,4,5,0]
```

Place stations at the position that maximizes future coverage:
- If the weak city is **not at the edge**: Place stations at the rightmost position in its range.
- If the weak city is **at the right edge**: Place stations at the leftmost position that can still reach it.

## Testing

```bash
go test -v
```

## Running

```bash
go run .
```