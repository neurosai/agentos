package catalog

import "time"

// Kind is a catalog entity type.
type Kind string

const (
	KindOrganization Kind = "Organization"
	KindTeam         Kind = "Team"
	KindSystem       Kind = "System"
	KindComponent    Kind = "Component"
	KindAPI          Kind = "API"
	KindResource     Kind = "Resource"
	KindRepository   Kind = "Repository"
	KindEnvironment  Kind = "Environment"
	KindCluster      Kind = "Cluster"
	KindNamespace    Kind = "Namespace"
	KindDataset      Kind = "Dataset"
	KindTool         Kind = "Tool"
	KindMCPServer    Kind = "MCPServer"
	KindAgent        Kind = "Agent"
)

// RelationType describes a directed edge in the catalog graph.
type RelationType string

const (
	RelationOwnedBy        RelationType = "ownedBy"
	RelationOwnerOf        RelationType = "ownerOf"
	RelationPartOf         RelationType = "partOf"
	RelationHasPart        RelationType = "hasPart"
	RelationDependsOn      RelationType = "dependsOn"
	RelationDependencyOf   RelationType = "dependencyOf"
	RelationProvidesAPI    RelationType = "providesApi"
	RelationConsumesAPI    RelationType = "consumesApi"
	RelationDeployedTo     RelationType = "deployedTo"
	RelationUsesResource   RelationType = "usesResource"
	RelationObservedCalls  RelationType = "observedCalls"
	RelationProducedByTool RelationType = "producedByTool"
	RelationExposedBy      RelationType = "exposedBy"
)

// Entity is a typed catalog node.
type Entity struct {
	Ref        string
	TenantID   string
	Kind       Kind
	Namespace  string
	Name       string
	Title      string
	Labels     map[string]string
	Spec       map[string]any
	Source     string // declared | observed
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Relation is a directed edge between entity refs.
type Relation struct {
	FromRef      string
	ToRef        string
	Type         RelationType
	TenantID     string
	Confidence   float64
	ObservedAt   *time.Time
	Declared     bool
}

// Snapshot is a point-in-time graph export.
type Snapshot struct {
	ID        string
	TenantID  string
	RootRef   string
	CreatedAt time.Time
	EntityIDs []string
	EdgeCount int
}

// GraphSlice is a bounded subgraph view.
type GraphSlice struct {
	Root     Entity
	Entities []Entity
	Edges    []Relation
}
