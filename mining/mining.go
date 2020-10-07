package mining

//HashRateReport is sent from the mining routines for giving combined information as output
type HashRateReport struct {
	MinerID  int
	HashRate float64
}

//Miner declares the common 'Mine' method
type Miner interface {
	Mine()
}
