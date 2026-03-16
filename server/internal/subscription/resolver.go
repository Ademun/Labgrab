package subscription

import (
	"labgrab/internal/shared/domain"
	"math"

	"github.com/google/uuid"
)

func (s *Service) ResolveConflicts(constraints map[uuid.UUID][]domain.Lesson) map[uuid.UUID][]domain.Lesson {
	graph, uuidMap, lessonMap := buildGraph(constraints)
	mcmf(graph, 0, len(graph)-1)

	uuidLessons := make(map[uuid.UUID][]domain.Lesson)
	for v, edges := range graph {
		if lesson, ok := lessonMap[v]; ok {
			for _, edge := range edges {
				if edge.Cap == 0 {
					if id, ok := uuidMap[edge.To]; ok {
						uuidLessons[id] = append(uuidLessons[id], lesson)
					}
				}
			}
		}
	}

	return uuidLessons
}

func addEdge(graph [][]Edge, from, to, cap, cost int) {
	graph[from] = append(graph[from], Edge{
		To:   to,
		Cap:  cap,
		Cost: cost,
		Rev:  len(graph[to]),
	})
	graph[to] = append(graph[to], Edge{
		To:   from,
		Cap:  0,
		Cost: -cost,
		Rev:  len(graph[from]) - 1,
	})
}

func spfa(graph [][]Edge, source int) ([]int, []int, []int) {
	dist := make([]int, len(graph))
	for i := range graph {
		dist[i] = math.MaxInt32
	}
	dist[source] = 0

	prevv := make([]int, len(graph))
	preve := make([]int, len(graph))
	queue := []int{source}
	inQueue := make([]bool, len(graph))
	inQueue[source] = true

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		inQueue[u] = false

		for i, edge := range graph[u] {
			if edge.Cap > 0 && dist[u]+edge.Cost < dist[edge.To] {
				dist[edge.To] = dist[u] + edge.Cost
				prevv[edge.To] = u
				preve[edge.To] = i
				if !inQueue[edge.To] {
					queue = append(queue, edge.To)
					inQueue[edge.To] = true
				}
			}
		}
	}

	return dist, prevv, preve
}

func augment(graph [][]Edge, source, sink int, dist, prevv, preve []int) int {
	v := sink
	minCap := math.MaxInt32
	for v != source {
		if e := graph[prevv[v]][preve[v]]; e.Cap < minCap {
			minCap = e.Cap
		}
		v = prevv[v]
	}
	v = sink
	for v != source {
		graph[prevv[v]][preve[v]].Cap -= minCap
		e := graph[prevv[v]][preve[v]]
		graph[e.To][e.Rev].Cap += minCap
		v = prevv[v]
	}
	return dist[sink] * minCap
}

func mcmf(graph [][]Edge, source, sink int) {
	for {
		dist, prevv, preve := spfa(graph, source)
		if dist[sink] == math.MaxInt32 {
			return
		}
		augment(graph, source, sink, dist, prevv, preve)
	}
}

func buildGraph(constraints map[uuid.UUID][]domain.Lesson) ([][]Edge, map[int]uuid.UUID, map[int]domain.Lesson) {
	rightSide := make(map[domain.Lesson]struct{})
	for _, lessons := range constraints {
		for _, l := range lessons {
			rightSide[l] = struct{}{}
		}
	}
	uuidMap := make(map[int]uuid.UUID, len(constraints))
	lessonMap := make(map[int]domain.Lesson, len(rightSide))
	reverseUUIDMap := make(map[uuid.UUID]int, len(constraints))
	reverseLessonMap := make(map[domain.Lesson]int, len(rightSide))
	maxLessons := (len(rightSide) + len(constraints) - 1) / len(constraints)

	idx := 1
	for lesson := range rightSide {
		lessonMap[idx] = lesson
		reverseLessonMap[lesson] = idx
		idx++
	}
	for id := range constraints {
		uuidMap[idx] = id
		reverseUUIDMap[id] = idx
		idx++
	}

	graph := make([][]Edge, len(constraints)+len(rightSide)+2)
	for idx := range lessonMap {
		addEdge(graph, 0, idx, 1, 0)
	}
	for idx := range uuidMap {
		addEdge(graph, idx, len(graph)-1, 1, 0)
		addEdge(graph, idx, len(graph)-1, maxLessons-1, 1)
	}
	for id, lessons := range constraints {
		for _, lesson := range lessons {
			addEdge(graph, reverseLessonMap[lesson], reverseUUIDMap[id], 1, 0)
		}
	}

	return graph, uuidMap, lessonMap
}
