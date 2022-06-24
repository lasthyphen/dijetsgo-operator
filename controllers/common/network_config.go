/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/lasthyphen/dijetsgo/ids"
	"github.com/lasthyphen/dijetsgo/utils/constants"
	"github.com/lasthyphen/dijetsgo/utils/hashing"
)

type Network struct {
	Genesis  string
	KeyPairs []KeyPair
}

type KeyPair struct {
	Cert string
	Key  string
	Id   string
}

func NewNetwork(networkSize int) (Network, error) {
	var (
		g Genesis
		n Network
	)
	if err := json.Unmarshal([]byte(defaultGenesisConfigJSON), &g); err != nil {
		return Network{}, fmt.Errorf("couldn't unmarshal local genesis: %w", err)
	}
	for i := 0; i < networkSize; i++ {
		// TODO handle the below error
		stakingKeyCertPair, _ := newStakingKeyCertPair()
		fmt.Print("------------------------------------------")
		fmt.Print(stakingKeyCertPair.Cert)
		fmt.Print("------------------------------------------")
		fmt.Print(stakingKeyCertPair.Key)
		fmt.Print("------------------------------------------")
		fmt.Print(stakingKeyCertPair.Id)
		n.KeyPairs = append(n.KeyPairs, stakingKeyCertPair)
		g.InitialStakers = append(g.InitialStakers, InitialStaker{NodeID: stakingKeyCertPair.Id, RewardAddress: g.Allocations[1].DjtxAddr, DelegationFee: 5000})
	}
	genesisBytes, err := json.Marshal(g)
	if err != nil {
		panic("Error: cannot marshal genesis.json, common package is invalid")
	}
	n.Genesis = string(genesisBytes)

	fmt.Print("------------------------------------------")
	fmt.Print(n.Genesis)

	return n, nil
}

func newStakingKeyCertPair() (KeyPair, error) {
	// Create key to sign cert with
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return KeyPair{}, fmt.Errorf("couldn't generate rsa key: %w", err)
	}

	// Create self-signed staking cert
	certTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(0),
		NotBefore:             time.Date(2020, time.January, 0, 0, 0, 0, 0, time.UTC),
		NotAfter:              time.Now().AddDate(100, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment,
		BasicConstraintsValid: true,
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &key.PublicKey, key)
	if err != nil {
		return KeyPair{}, fmt.Errorf("couldn't create certificate: %w", err)
	}

	var certBuff bytes.Buffer
	if err := pem.Encode(&certBuff, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return KeyPair{}, fmt.Errorf("couldn't write cert file: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return KeyPair{}, fmt.Errorf("couldn't marshal private key: %w", err)
	}

	var keyBuff bytes.Buffer
	if err := pem.Encode(&keyBuff, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return KeyPair{}, fmt.Errorf("couldn't write private key: %w", err)
	}

	id, err := ids.ToShortID(hashing.PubkeyBytesToAddress(certBytes))
	if err != nil {
		return KeyPair{}, fmt.Errorf("problem deriving node ID from certificate: %w", err)
	}
	fullId := id.PrefixedString(constants.NodeIDPrefix)

	return KeyPair{
		Cert: certBuff.String(),
		Key:  keyBuff.String(),
		Id:   fullId,
	}, nil
}
