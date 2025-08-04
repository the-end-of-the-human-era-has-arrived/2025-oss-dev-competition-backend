package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/repository"
)

func TestGetMindMapByUser(t *testing.T) {
	tests := []struct {
		name               string
		setupData          func(*repository.MemoryMindMapRepo) uuid.UUID
		expectedNodesCount int
		validateResult     func(*testing.T, *domain.MindMapGraph, uuid.UUID)
	}{
		{
			name: "Empty case - user has no nodes",
			setupData: func(repo *repository.MemoryMindMapRepo) uuid.UUID {
				userID := uuid.New()
				// Create nodes for different user to ensure proper filtering
				otherUserID := uuid.New()
				notionID := uuid.New()
				repo.CreateKeywordNode(context.Background(), otherUserID, notionID, "other user node")
				return userID
			},
			expectedNodesCount: 0,
			validateResult: func(t *testing.T, result *domain.MindMapGraph, userID uuid.UUID) {
				if result.UserID != userID {
					t.Errorf("Expected UserID %v, got %v", userID, result.UserID)
				}
				if len(result.AdjacencyList) != 0 {
					t.Errorf("Expected empty adjacency list, got %d entries", len(result.AdjacencyList))
				}
			},
		},
		{
			name: "Single isolated node",
			setupData: func(repo *repository.MemoryMindMapRepo) uuid.UUID {
				userID := uuid.New()
				notionID := uuid.New()
				repo.CreateKeywordNode(context.Background(), userID, notionID, "single node")
				return userID
			},
			expectedNodesCount: 1,
			validateResult: func(t *testing.T, result *domain.MindMapGraph, userID uuid.UUID) {
				if len(result.AdjacencyList) != 1 {
					t.Errorf("Expected 1 node in adjacency list, got %d", len(result.AdjacencyList))
				}
				for nodeID, neighbors := range result.AdjacencyList {
					if len(neighbors) != 0 {
						t.Errorf("Expected no neighbors for isolated node %v, got %d", nodeID, len(neighbors))
					}
				}
			},
		},
		{
			name: "Multiple isolated nodes",
			setupData: func(repo *repository.MemoryMindMapRepo) uuid.UUID {
				userID := uuid.New()
				notionID := uuid.New()
				repo.CreateKeywordNode(context.Background(), userID, notionID, "node1")
				repo.CreateKeywordNode(context.Background(), userID, notionID, "node2")
				repo.CreateKeywordNode(context.Background(), userID, notionID, "node3")
				return userID
			},
			expectedNodesCount: 3,
			validateResult: func(t *testing.T, result *domain.MindMapGraph, userID uuid.UUID) {
				if len(result.AdjacencyList) != 3 {
					t.Errorf("Expected 3 nodes in adjacency list, got %d", len(result.AdjacencyList))
				}
				for nodeID, neighbors := range result.AdjacencyList {
					if len(neighbors) != 0 {
						t.Errorf("Expected no neighbors for isolated node %v, got %d", nodeID, len(neighbors))
					}
				}
			},
		},
		{
			name: "Two connected nodes",
			setupData: func(repo *repository.MemoryMindMapRepo) uuid.UUID {
				userID := uuid.New()
				notionID := uuid.New()
				node1, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "node1")
				node2, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "node2")
				repo.CreateKeywordEdge(context.Background(), node1, node2)
				return userID
			},
			expectedNodesCount: 2,
			validateResult: func(t *testing.T, result *domain.MindMapGraph, userID uuid.UUID) {
				if len(result.AdjacencyList) != 2 {
					t.Errorf("Expected 2 nodes in adjacency list, got %d", len(result.AdjacencyList))
				}
				
				// Each node should have exactly 1 neighbor
				for nodeID, neighbors := range result.AdjacencyList {
					if len(neighbors) != 1 {
						t.Errorf("Expected 1 neighbor for node %v, got %d", nodeID, len(neighbors))
					}
				}
			},
		},
		{
			name: "Chain of nodes (A-B-C)",
			setupData: func(repo *repository.MemoryMindMapRepo) uuid.UUID {
				userID := uuid.New()
				notionID := uuid.New()
				nodeA, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "nodeA")
				nodeB, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "nodeB")
				nodeC, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "nodeC")
				
				repo.CreateKeywordEdge(context.Background(), nodeA, nodeB)
				repo.CreateKeywordEdge(context.Background(), nodeB, nodeC)
				return userID
			},
			expectedNodesCount: 3,
			validateResult: func(t *testing.T, result *domain.MindMapGraph, userID uuid.UUID) {
				if len(result.AdjacencyList) != 3 {
					t.Errorf("Expected 3 nodes in adjacency list, got %d", len(result.AdjacencyList))
				}
				
				// Count nodes by neighbor count
				neighborCounts := make(map[int]int)
				for _, neighbors := range result.AdjacencyList {
					neighborCounts[len(neighbors)]++
				}
				
				// In a chain A-B-C: 2 nodes have 1 neighbor (A,C) and 1 node has 2 neighbors (B)
				if neighborCounts[1] != 2 {
					t.Errorf("Expected 2 nodes with 1 neighbor, got %d", neighborCounts[1])
				}
				if neighborCounts[2] != 1 {
					t.Errorf("Expected 1 node with 2 neighbors, got %d", neighborCounts[2])
				}
			},
		},
		{
			name: "Star topology (center connected to 3 nodes)",
			setupData: func(repo *repository.MemoryMindMapRepo) uuid.UUID {
				userID := uuid.New()
				notionID := uuid.New()
				center, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "center")
				node1, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "node1")
				node2, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "node2")
				node3, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "node3")
				
				repo.CreateKeywordEdge(context.Background(), center, node1)
				repo.CreateKeywordEdge(context.Background(), center, node2)
				repo.CreateKeywordEdge(context.Background(), center, node3)
				return userID
			},
			expectedNodesCount: 4,
			validateResult: func(t *testing.T, result *domain.MindMapGraph, userID uuid.UUID) {
				if len(result.AdjacencyList) != 4 {
					t.Errorf("Expected 4 nodes in adjacency list, got %d", len(result.AdjacencyList))
				}
				
				// Count nodes by neighbor count
				neighborCounts := make(map[int]int)
				for _, neighbors := range result.AdjacencyList {
					neighborCounts[len(neighbors)]++
				}
				
				// In star topology: 3 leaf nodes have 1 neighbor, 1 center has 3 neighbors
				if neighborCounts[1] != 3 {
					t.Errorf("Expected 3 nodes with 1 neighbor, got %d", neighborCounts[1])
				}
				if neighborCounts[3] != 1 {
					t.Errorf("Expected 1 node with 3 neighbors, got %d", neighborCounts[3])
				}
			},
		},
		{
			name: "Complex graph with cycle",
			setupData: func(repo *repository.MemoryMindMapRepo) uuid.UUID {
				userID := uuid.New()
				notionID := uuid.New()
				nodeA, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "nodeA")
				nodeB, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "nodeB")
				nodeC, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "nodeC")
				nodeD, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "nodeD")
				
				// Create a cycle A-B-C-A and connect D to A
				repo.CreateKeywordEdge(context.Background(), nodeA, nodeB)
				repo.CreateKeywordEdge(context.Background(), nodeB, nodeC)
				repo.CreateKeywordEdge(context.Background(), nodeC, nodeA)
				repo.CreateKeywordEdge(context.Background(), nodeA, nodeD)
				return userID
			},
			expectedNodesCount: 4,
			validateResult: func(t *testing.T, result *domain.MindMapGraph, userID uuid.UUID) {
				if len(result.AdjacencyList) != 4 {
					t.Errorf("Expected 4 nodes in adjacency list, got %d", len(result.AdjacencyList))
				}
				
				// Count total edges (each edge should be counted twice in adjacency list)
				totalEdges := 0
				for _, neighbors := range result.AdjacencyList {
					totalEdges += len(neighbors)
				}
				expectedTotalEdges := 4 * 2 // 4 edges, each counted twice
				if totalEdges != expectedTotalEdges {
					t.Errorf("Expected %d total edge entries, got %d", expectedTotalEdges, totalEdges)
				}
			},
		},
		{
			name: "Mixed users - filter correctly",
			setupData: func(repo *repository.MemoryMindMapRepo) uuid.UUID {
				userID := uuid.New()
				otherUserID := uuid.New()
				notionID := uuid.New()
				
				// Create nodes for target user
				targetNode1, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "target1")
				targetNode2, _ := repo.CreateKeywordNode(context.Background(), userID, notionID, "target2")
				
				// Create nodes for other users
				otherNode1, _ := repo.CreateKeywordNode(context.Background(), otherUserID, notionID, "other1")
				otherNode2, _ := repo.CreateKeywordNode(context.Background(), otherUserID, notionID, "other2")
				
				// Connect target user nodes
				repo.CreateKeywordEdge(context.Background(), targetNode1, targetNode2)
				
				// Connect other user nodes (should not appear in result)
				repo.CreateKeywordEdge(context.Background(), otherNode1, otherNode2)
				
				// Cross-user connection (should not appear in result)
				repo.CreateKeywordEdge(context.Background(), targetNode1, otherNode1)
				
				return userID
			},
			expectedNodesCount: 2,
			validateResult: func(t *testing.T, result *domain.MindMapGraph, userID uuid.UUID) {
				if len(result.AdjacencyList) != 2 {
					t.Errorf("Expected 2 nodes in adjacency list, got %d", len(result.AdjacencyList))
				}
				
				// Verify that nodes belong to the correct user
				if result.UserID != userID {
					t.Errorf("Expected UserID %v, got %v", userID, result.UserID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			repo := repository.NewMemoryMindMapRepo()
			usecase := NewMindMapUsecase(repo)
			userID := tt.setupData(repo)
			
			// Execute
			result := usecase.GetMindMapByUser(context.Background(), userID)
			
			// Validate
			if result == nil {
				t.Fatal("Expected non-nil result")
			}
			
			if result.UserID != userID {
				t.Errorf("Expected UserID %v, got %v", userID, result.UserID)
			}
			
			if len(result.AdjacencyList) != tt.expectedNodesCount {
				t.Errorf("Expected %d nodes, got %d", tt.expectedNodesCount, len(result.AdjacencyList))
			}
			
			// Run custom validation
			tt.validateResult(t, result, userID)
		})
	}
}

// Test helper functions
func TestBuildNeighborList(t *testing.T) {
	nodeID := uuid.New()
	neighbor1 := uuid.New()
	neighbor2 := uuid.New()
	neighbor3 := uuid.New()
	
	edges := []*domain.KeywordEdge{
		{ID: uuid.New(), Keyword1: nodeID, Keyword2: neighbor1},
		{ID: uuid.New(), Keyword1: neighbor2, Keyword2: nodeID},
		{ID: uuid.New(), Keyword1: nodeID, Keyword2: neighbor3},
	}
	
	result := buildNeighborList(nodeID, edges)
	
	if len(result) != 3 {
		t.Errorf("Expected 3 neighbors, got %d", len(result))
	}
	
	// Verify all expected neighbors are present
	expectedNeighbors := map[uuid.UUID]bool{
		neighbor1: false,
		neighbor2: false,
		neighbor3: false,
	}
	
	for _, neighbor := range result {
		if _, exists := expectedNeighbors[neighbor]; exists {
			expectedNeighbors[neighbor] = true
		} else {
			t.Errorf("Unexpected neighbor %v", neighbor)
		}
	}
	
	// Check all expected neighbors were found
	for neighbor, found := range expectedNeighbors {
		if !found {
			t.Errorf("Expected neighbor %v not found", neighbor)
		}
	}
}

func TestGetNeighborIDFromEdge(t *testing.T) {
	nodeA := uuid.New()
	nodeB := uuid.New()
	
	edge := &domain.KeywordEdge{
		ID:       uuid.New(),
		Keyword1: nodeA,
		Keyword2: nodeB,
	}
	
	// Test getting neighbor when node is Keyword1
	neighbor := getNeighborIDFromEdge(nodeA, edge)
	if neighbor != nodeB {
		t.Errorf("Expected neighbor %v, got %v", nodeB, neighbor)
	}
	
	// Test getting neighbor when node is Keyword2
	neighbor = getNeighborIDFromEdge(nodeB, edge)
	if neighbor != nodeA {
		t.Errorf("Expected neighbor %v, got %v", nodeA, neighbor)
	}
}