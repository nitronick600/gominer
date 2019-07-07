package equihash

import (
	"sync"

	"github.com/robvanmieghem/gominer/clients"
	"github.com/robvanmieghem/gominer/clients/stratum"
)

const (
	//HashSize is the length of an equi hash
	HashSize = 32
)

//Target declares what a solution should be smaller than to be accepted
type Target [HashSize]byte

type equihashJob struct {
}

type EquihashClient struct {
	connectionstring string
	User             string
	Password         string

	*sync.Mutex     // protects following
	stratumclient   *stratum.Client
	extranonce1     []byte
	extranonce2Size uint
	target          Target
	currentJob      equihashJob
	clients.BaseClient
}
