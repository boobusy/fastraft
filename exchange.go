package fastraft

type Exchange interface {
	Replication() bool
	Vote()
}
