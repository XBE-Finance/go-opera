/*
Copyright 2017 Mosaic Networks Ltd

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
package node

import (
	"math/rand"
	"sort"

	"bitbucket.org/mosaicnet/babble/net"
)

type PeerSelector interface {
	Peers() []net.Peer
	Next() net.Peer
}

//+++++++++++++++++++++++++++++++++++++++
//DETERMINISTIC

type DeterministicPeerSelector struct {
	peers []net.Peer
	index int
}

func NewDeterministicPeerSelector(participants []net.Peer, localAddr string) *DeterministicPeerSelector {
	sort.Sort(net.ByPubKey(participants))
	position, peers := net.ExcludePeer(participants, localAddr)
	index := position % len(peers)
	return &DeterministicPeerSelector{
		peers: peers,
		index: index,
	}
}

func (ps *DeterministicPeerSelector) Peers() []net.Peer {
	return ps.peers
}

func (ps *DeterministicPeerSelector) Next() net.Peer {
	next := ps.peers[ps.index]
	ps.index = (ps.index + 1) % len(ps.peers)
	return next
}

//+++++++++++++++++++++++++++++++++++++++
//RANDOM

type RandomPeerSelector struct {
	peers []net.Peer
}

func NewRandomPeerSelector(participants []net.Peer, localAddr string) *RandomPeerSelector {
	_, peers := net.ExcludePeer(participants, localAddr)
	return &RandomPeerSelector{
		peers: peers,
	}
}

func (ps *RandomPeerSelector) Peers() []net.Peer {
	return ps.peers
}

func (ps *RandomPeerSelector) Next() net.Peer {
	i := rand.Intn(len(ps.peers))
	peer := ps.peers[i]
	return peer
}