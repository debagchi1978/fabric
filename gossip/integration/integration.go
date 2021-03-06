/*
Copyright IBM Corp. 2016 All Rights Reserved.

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

package integration

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/gossip/comm"
	"github.com/hyperledger/fabric/gossip/gossip"
	"github.com/hyperledger/fabric/gossip/proto"
	"google.golang.org/grpc"
	"github.com/hyperledger/fabric/gossip/api"
	"github.com/hyperledger/fabric/gossip/common"
)

// This file is used to bootstrap a gossip instance for integration/demo purposes ONLY

func newConfig(selfEndpoint string, bootPeers ...string) *gossip.Config {
	port, err := strconv.ParseInt(strings.Split(selfEndpoint, ":")[1], 10, 64)
	if err != nil {
		panic(err)
	}
	return &gossip.Config{
		BindPort:       int(port),
		BootstrapPeers: bootPeers,
		ID:             selfEndpoint,
		MaxMessageCountToStore:     100,
		MaxPropagationBurstLatency: time.Millisecond * 50,
		MaxPropagationBurstSize:    3,
		PropagateIterations:        1,
		PropagatePeerNum:           3,
		PullInterval:               time.Second * 5,
		PullPeerNum:                3,
		SelfEndpoint:               selfEndpoint,
	}
}

func newComm(selfEndpoint string, s *grpc.Server, dialOpts ...grpc.DialOption) comm.Comm {
	comm, err := comm.NewCommInstance(s, NewGossipCryptoService(), []byte(selfEndpoint), dialOpts...)
	if err != nil {
		panic(err)
	}
	return comm
}

// NewGossipComponent creates a gossip component that attaches itself to the given gRPC server
func NewGossipComponent(endpoint string, s *grpc.Server, bootPeers ...string) (gossip.Gossip, comm.Comm) {
	conf := newConfig(endpoint, bootPeers...)
	comm := newComm(endpoint, s, grpc.WithInsecure())
	return gossip.NewGossipService(conf, comm, &naiveCryptoService{}, NewGossipCryptoService(), api.PeerIdentityType(conf.ID)), comm
}

// GossipCryptoService is an interface that conforms to both
// the comm.SecurityProvider and to discovery.CryptoService
type GossipCryptoService interface {

	// isEnabled returns whether authentication is enabled
	IsEnabled() bool

	// Sign signs msg with this peers signing key and outputs
	// the signature if no error occurred.
	Sign(msg []byte) ([]byte, error)

	// Verify checks that signature if a valid signature of message under vkID's verification key.
	// If the verification succeeded, Verify returns nil meaning no error occurred.
	// If vkID is nil, then the signature is verified against this validator's verification key.
	Verify(vkID, signature, message []byte) error

	// validateAliveMsg validates that an Alive message is authentic
	ValidateAliveMsg(*proto.AliveMessage) bool

	// SignMessage signs an AliveMessage and updates its signature field
	SignMessage(*proto.AliveMessage) *proto.AliveMessage
}

// NewGossipCryptoService returns an instance that implements naively every security
// interface that the gossip layer needs
func NewGossipCryptoService() GossipCryptoService {
	return &naiveCryptoServiceImpl{}
}

type naiveCryptoServiceImpl struct {
}

func (cs *naiveCryptoServiceImpl) ValidateAliveMsg(*proto.AliveMessage) bool {
	return true
}

func (cs *naiveCryptoServiceImpl) Verify(vkID, signature, message []byte) error {
	if ! bytes.Equal(signature, message) {
		return fmt.Errorf("Wrong signature")
	}
	return nil
}

// SignMessage signs an AliveMessage and updates its signature field
func (cs *naiveCryptoServiceImpl) SignMessage(msg *proto.AliveMessage) *proto.AliveMessage {
	return msg
}

// IsEnabled returns true whether authentication is enabled
func (cs *naiveCryptoServiceImpl) IsEnabled() bool {
	return false
}

// Sign signs a message with the local peer's private key
func (cs *naiveCryptoServiceImpl) Sign(msg []byte) ([]byte, error) {
	return msg, nil
}

// Verify checks that signature is a valid signature of message under a peer's verification key.
// If the verification succeeded, Verify returns nil meaning no error occurred.
// If peerCert is nil, then the signature is verified against this peer's verification key.
func (*naiveCryptoService) Verify(peerIdentity api.PeerIdentityType, signature, message []byte) error {
	equal := bytes.Equal(signature, message)
	if !equal {
		return fmt.Errorf("Wrong signature:%v, %v", signature, message)
	}
	return nil
}

type naiveCryptoService struct {
}

func (*naiveCryptoService) ValidateAliveMsg(am *proto.AliveMessage) bool {
	return true
}

func (*naiveCryptoService) SignMessage(am *proto.AliveMessage) *proto.AliveMessage {
	return am
}

func (*naiveCryptoService) IsEnabled() bool {
	return true
}

func (*naiveCryptoService) ValidateIdentity(peerIdentity api.PeerIdentityType) error {
	return nil
}

// GetPKIidOfCert returns the PKI-ID of a peer's identity
func (*naiveCryptoService) GetPKIidOfCert(peerIdentity api.PeerIdentityType) common.PKIidType {
	return common.PKIidType(peerIdentity)
}

// VerifyBlock returns nil if the block is properly signed,
// else returns error
func (*naiveCryptoService) VerifyBlock(signedBlock api.SignedBlock) error {
	return nil
}

// Sign signs msg with this peer's signing key and outputs
// the signature if no error occurred.
func (*naiveCryptoService) Sign(msg []byte) ([]byte, error) {
	return msg, nil
}