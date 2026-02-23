package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	v32 "github.com/blcvn/backend/services/pkg/domain/v3.2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NodeModel represents a node in the requirement graph
type NodeModel struct {
	ID          string `gorm:"primaryKey"`
	ProjectID   string `gorm:"index"`
	Type        string `gorm:"index"`
	Summary     string
	Description string
	SourceID    string
	Metadata    []byte `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (NodeModel) TableName() string {
	return "graph_nodes"
}

// EdgeModel represents an edge in the requirement graph
type EdgeModel struct {
	SourceID  string `gorm:"primaryKey"`
	TargetID  string `gorm:"primaryKey"`
	ProjectID string `gorm:"index"`
	Type      string `gorm:"index"`
	Reason    string
	CreatedAt time.Time
}

func (EdgeModel) TableName() string {
	return "graph_edges"
}

// GraphRepository implements v32.GraphRepository
type GraphRepository struct {
	db *gorm.DB
}

func NewGraphRepository(db *gorm.DB) *GraphRepository {
	return &GraphRepository{db: db}
}

func (r *GraphRepository) Save(ctx context.Context, docID string, graph *v32.RequirementGraph) error {
	// Debug Log
	fmt.Printf("[GraphRepository] Saving Graph for doc: %s. Nodes: %d, Edges: %d\n", docID, len(graph.Nodes), len(graph.Edges))

	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete existing graph data for this document
		// We delete nodes and edges to ensure we have a clean state for this document version
		if err := tx.Where("document_id = ?", docID).Delete(&v32.RequirementNode{}).Error; err != nil {
			return err
		}
		if err := tx.Where("document_id = ?", docID).Delete(&v32.DependencyEdge{}).Error; err != nil {
			return err
		}
		// We DO NOT delete the graph metadata here to avoid race conditions with checking existence.
		// Instead we use Upsert on creation.

		// 2. Set IDs and DocumentIDs if missing
		graph.DocumentID = docID
		if graph.ID == "" {
			graph.ID = fmt.Sprintf("graph-%s", docID) // Simple ID strategy
		}

		// Ensure uniqueness map
		uniqueNodes := make(map[string]bool)
		validNodes := make([]v32.RequirementNode, 0, len(graph.Nodes))

		for i := range graph.Nodes {
			node := graph.Nodes[i]
			node.DocumentID = docID
			if node.ID == "" {
				node.ID = fmt.Sprintf("node-%s-%d", docID, i)
			}

			if !uniqueNodes[node.ID] {
				uniqueNodes[node.ID] = true
				validNodes = append(validNodes, node)
			} else {
				fmt.Printf("[GraphRepository] Warning: Skipping duplicate node ID in Save: %s\n", node.ID)
			}
		}
		graph.Nodes = validNodes // Update graph nodes to valid only

		uniqueEdges := make(map[string]bool)
		validEdges := make([]v32.DependencyEdge, 0, len(graph.Edges))

		for i := range graph.Edges {
			edge := graph.Edges[i]
			edge.DocumentID = docID
			if edge.ID == "" {
				edge.ID = fmt.Sprintf("edge-%s-%d", docID, i)
			}

			if !uniqueEdges[edge.ID] {
				uniqueEdges[edge.ID] = true
				validEdges = append(validEdges, edge)
			}
		}
		graph.Edges = validEdges

		// 3. Insert or Update Graph Metadata (Upsert)
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"document_id", "metadata"}),
		}).Create(graph).Error; err != nil {
			return fmt.Errorf("failed to save graph metadata: %w", err)
		}

		// 4. Insert Nodes with ON CONFLICT DO UPDATE to be safe against phantom reads or concurrent writes
		if len(graph.Nodes) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"type", "summary", "description", "document_id", "reference_id", "source_id", "metadata"}),
			}).Create(&graph.Nodes).Error; err != nil {
				return fmt.Errorf("failed to save nodes: %w", err)
			}
		}

		// 5. Insert Edges with DO NOTHING (or Update)
		if len(graph.Edges) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"type", "reason", "document_id", "source_id", "target_id"}),
			}).Create(&graph.Edges).Error; err != nil {
				return fmt.Errorf("failed to save edges: %w", err)
			}
		}

		return nil
	})
}

func (r *GraphRepository) GetByDocumentID(ctx context.Context, docID string) (*v32.RequirementGraph, error) {
	fmt.Printf("[GraphRepository] Getting Graph for doc: %s\n", docID)
	var graph v32.RequirementGraph

	// Get Graph Metadata
	if err := r.db.Where("document_id = ?", docID).First(&graph).Error; err != nil {
		fmt.Printf("[GraphRepository] Graph metadata not found for doc: %s. Error: %v\n", docID, err)
		return nil, err
	}

	// Get Nodes
	if err := r.db.Where("document_id = ?", docID).Find(&graph.Nodes).Error; err != nil {
		fmt.Printf("[GraphRepository] Failed to find nodes for doc: %s. Error: %v\n", docID, err)
		return nil, err
	}

	// Get Edges
	if err := r.db.Where("document_id = ?", docID).Find(&graph.Edges).Error; err != nil {
		fmt.Printf("[GraphRepository] Failed to find edges for doc: %s. Error: %v\n", docID, err)
		return nil, err
	}

	fmt.Printf("[GraphRepository] Found Graph for doc: %s. Nodes: %d, Edges: %d\n", docID, len(graph.Nodes), len(graph.Edges))
	return &graph, nil
}

// SaveGraph saves a full graph (nodes and edges) transactionally
func (r *GraphRepository) SaveGraph(ctx context.Context, projectID string, graph *v32.RequirementGraph) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Process Nodes
		if len(graph.Nodes) > 0 {
			var nodeModels []NodeModel
			for _, n := range graph.Nodes {
				meta, _ := json.Marshal(n.Metadata)
				nodeModels = append(nodeModels, NodeModel{
					ID:          n.ID,
					ProjectID:   projectID,
					Type:        string(n.Type),
					Summary:     n.Summary,
					Description: n.Description,
					SourceID:    n.SourceID,
					Metadata:    meta,
				})
			}
			// Upsert nodes
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"type", "summary", "description", "source_id", "metadata", "updated_at"}),
			}).Create(&nodeModels).Error; err != nil {
				return err
			}
		}

		// 2. Process Edges
		if len(graph.Edges) > 0 {
			var edgeModels []EdgeModel
			for _, e := range graph.Edges {
				edgeModels = append(edgeModels, EdgeModel{
					SourceID:  e.SourceID,
					TargetID:  e.TargetID,
					ProjectID: projectID,
					Type:      string(e.Type),
					Reason:    e.Reason,
				})
			}
			// Upsert edges
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "source_id"}, {Name: "target_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"type", "reason"}),
			}).Create(&edgeModels).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetGraph retrieves the full graph for a project
func (r *GraphRepository) GetGraph(ctx context.Context, projectID string) (*v32.RequirementGraph, error) {
	var nodeModels []NodeModel
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&nodeModels).Error; err != nil {
		return nil, err
	}

	var edgeModels []EdgeModel
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&edgeModels).Error; err != nil {
		return nil, err
	}

	graph := v32.NewRequirementGraph()
	for _, nm := range nodeModels {
		var meta map[string]any
		_ = json.Unmarshal(nm.Metadata, &meta)
		graph.AddNode(v32.RequirementNode{
			ID:          nm.ID,
			Type:        v32.RequirementType(nm.Type),
			Summary:     nm.Summary,
			Description: nm.Description,
			SourceID:    nm.SourceID,
			Metadata:    meta,
		})
	}

	for _, em := range edgeModels {
		graph.AddEdge(v32.DependencyEdge{
			SourceID: em.SourceID,
			TargetID: em.TargetID,
			Type:     v32.DependencyType(em.Type),
			Reason:   em.Reason,
		})
	}

	// Load graph metadata if needed (not persisted in this simple schema, typically project level)
	return graph, nil
}

// AddNode adds a single node
func (r *GraphRepository) AddNode(ctx context.Context, projectID string, node v32.RequirementNode) error {
	meta, _ := json.Marshal(node.Metadata)
	model := NodeModel{
		ID:          node.ID,
		ProjectID:   projectID,
		Type:        string(node.Type),
		Summary:     node.Summary,
		Description: node.Description,
		SourceID:    node.SourceID,
		Metadata:    meta,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"type", "summary", "description", "source_id", "metadata", "updated_at"}),
	}).Create(&model).Error
}

// AddEdge adds a single edge
func (r *GraphRepository) AddEdge(ctx context.Context, projectID string, edge v32.DependencyEdge) error {
	model := EdgeModel{
		SourceID:  edge.SourceID,
		TargetID:  edge.TargetID,
		ProjectID: projectID,
		Type:      string(edge.Type),
		Reason:    edge.Reason,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "source_id"}, {Name: "target_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"type", "reason"}),
	}).Create(&model).Error
}

// Verify interface compliance
var _ v32.GraphRepository = (*GraphRepository)(nil)
