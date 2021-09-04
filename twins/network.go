package twins

import (
	"context"
	"fmt"
	"sync"

	"github.com/relab/hotstuff"
	"github.com/relab/hotstuff/consensus"
)

type NodeID struct {
	ReplicaID hotstuff.ID
	NetworkID uint32
}

type Node struct {
	ID      NodeID
	Modules *consensus.Modules
}

type Network struct {
	Nodes map[NodeID]*Node
	// Maps a replica ID to a replica and its twins.
	Replicas map[hotstuff.ID][]*Node
	// For each view (starting at 1), contains the list of partitions for that view.
	Partitions [][]NodeSet

	mut sync.Mutex
	// The view in which the last timeout occurred for a node.
	lastTimeouts map[NodeID]consensus.View
	// hungNodes the set of nodes which have
	hungNodes NodeSet

	allHung chan struct{}
}

func NewNetwork(partitions [][]NodeSet) *Network {
	return &Network{
		Nodes:        make(map[NodeID]*Node),
		Replicas:     make(map[hotstuff.ID][]*Node),
		Partitions:   partitions,
		lastTimeouts: make(map[NodeID]consensus.View),
		hungNodes:    make(NodeSet),
		allHung:      make(chan struct{}),
	}
}

func (n *Network) WaitUntilHung() {
	<-n.allHung
}

func (n *Network) timeout(node NodeID, view consensus.View) {
	n.mut.Lock()
	defer n.mut.Unlock()

	lastTimeoutView, ok := n.lastTimeouts[node]

	if ok && lastTimeoutView == view {
		n.hungNodes.Add(node)
		if len(n.hungNodes) == len(n.Nodes) {
			close(n.allHung)
		}
		return
	}

	if lastTimeoutView > view {
		// strange, but we'll ignore it
		return
	}

	n.lastTimeouts[node] = view
	delete(n.hungNodes, node)
}

// ShouldDrop decides if the sender should drop the message, based on the current view of the sender and the
// partitions configured for that view.
func (n *Network) ShouldDrop(sender, receiver NodeID) bool {
	node, ok := n.Nodes[sender]
	if !ok {
		panic(fmt.Errorf("node matching sender id %d was not found", sender))
	}

	// Index into viewPartitions.
	i := int(node.Modules.Synchronizer().View() - 1)

	// will default to dropping all messages from views that don't have any specified partitions.
	if i >= len(n.Partitions) {
		return true
	}

	partitions := n.Partitions[i]
	for _, partition := range partitions {
		if partition.Contains(sender) && partition.Contains(receiver) {
			return false
		}
	}

	return true
}

type configuration struct {
	node    *Node
	network *Network
}

func (c *configuration) broadcastMessage(message interface{}) {
	for id := range c.network.Replicas {
		c.sendMessage(id, message)
	}
}

func (c *configuration) sendMessage(id hotstuff.ID, message interface{}) {
	nodes, ok := c.network.Replicas[id]
	if !ok {
		panic(fmt.Errorf("attempt to send message to replica %d, but this replica does not exist", id))
	}
	for _, node := range nodes {
		if c.shouldDrop(node.ID) {
			continue
		}
		node.Modules.EventLoop().AddEvent(message)
	}
}

// shouldDrop checks if a message to the node identified by id should be dropped.
func (c *configuration) shouldDrop(id NodeID) bool {
	// retrieve the drop config for this node.
	return c.network.ShouldDrop(c.node.ID, id)
}

// Replicas returns all of the replicas in the configuration.
func (c *configuration) Replicas() map[hotstuff.ID]consensus.Replica {
	m := make(map[hotstuff.ID]consensus.Replica)
	for id := range c.network.Replicas {
		m[id] = &replica{
			config: c,
			id:     id,
		}
	}
	return m
}

// Replica returns a replica if present in the configuration.
func (c *configuration) Replica(id hotstuff.ID) (r consensus.Replica, ok bool) {
	if _, ok = c.network.Replicas[id]; ok {
		return &replica{
			config: c,
			id:     id,
		}, true
	}
	return nil, false
}

// Len returns the number of replicas in the configuration.
func (c *configuration) Len() int {
	return len(c.network.Replicas)
}

// QuorumSize returns the size of a quorum.
func (c *configuration) QuorumSize() int {
	return hotstuff.QuorumSize(c.Len())
}

// Propose sends the block to all replicas in the configuration.
func (c *configuration) Propose(proposal consensus.ProposeMsg) {
	c.broadcastMessage(proposal)
}

// Timeout sends the timeout message to all replicas.
func (c *configuration) Timeout(msg consensus.TimeoutMsg) {
	c.broadcastMessage(msg)
}

// Fetch requests a block from all the replicas in the configuration.
func (c *configuration) Fetch(_ context.Context, hash consensus.Hash) (block *consensus.Block, ok bool) {
	for _, replica := range c.network.Replicas {
		for _, node := range replica {
			if c.shouldDrop(node.ID) {
				continue
			}
			block, ok = node.Modules.BlockChain().LocalGet(hash)
			if ok {
				return block, true
			}
		}
	}
	return nil, false
}

type replica struct {
	// pointer to the node that wants to contact this replica.
	config *configuration
	// id of the replica.
	id hotstuff.ID
}

// ID returns the replica's id.
func (r *replica) ID() hotstuff.ID {
	return r.config.network.Replicas[r.id][0].ID.ReplicaID
}

// PublicKey returns the replica's public key.
func (r *replica) PublicKey() consensus.PublicKey {
	return r.config.network.Replicas[r.id][0].Modules.PrivateKey().Public()
}

// Vote sends the partial certificate to the other replica.
func (r *replica) Vote(cert consensus.PartialCert) {
	r.config.sendMessage(r.id, consensus.VoteMsg{
		ID:          r.config.node.Modules.ID(),
		PartialCert: cert,
	})
}

// NewView sends the quorum certificate to the other replica.
func (r *replica) NewView(si consensus.SyncInfo) {
	r.config.sendMessage(r.id, consensus.NewViewMsg{
		ID:       r.config.node.Modules.ID(),
		SyncInfo: si,
	})
}

type NodeSet map[NodeID]struct{}

func (s NodeSet) Add(v NodeID) {
	s[v] = struct{}{}
}

func (s NodeSet) Contains(v NodeID) bool {
	_, ok := s[v]
	return ok
}
