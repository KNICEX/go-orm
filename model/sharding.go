package model

type ShardingAlgorithm interface {
	Sharding(sk any) (database string, table string)
}
