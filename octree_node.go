package octree

import (
	"fmt"
	"math"

	"github.com/The-Tensox/protometry"
)

type OctreeNode struct {
	position *protometry.VectorN
	data     []interface{}
	region   *protometry.Box
	children [8]*OctreeNode
}

func NewEmptyOctreeNode() *OctreeNode {
	return &OctreeNode{position: protometry.NewVectorN(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64)}
}

func NewPointOctreeNode(position *protometry.VectorN, data []interface{}) *OctreeNode {
	return &OctreeNode{position: position, data: data}
}

func NewRegionOctreeNode(region *protometry.Box) *OctreeNode {
	o := NewEmptyOctreeNode()
	o.position = nil
	o.region = region
	for i := TLF; i <= BLB; i++ {
		o.children[i] = NewEmptyOctreeNode()
	}
	return o
}

func (o *OctreeNode) Search(position protometry.VectorN) (*OctreeNode, error) {
	pos := o.getOctant(position)
	if o.children[pos].position == nil {
		// Region node
		return o.children[pos].Search(position)
	} else if o.children[pos].position.Dimensions[0] == math.MaxFloat64 {
		// Empty node
		return nil, ErrtreeFailedToFindNode
	}
	eq, err := o.children[pos].position.ApproxEqual(position)
	if err != nil {
		return nil, err
	}
	if eq {
		return o.children[pos], nil
	}
	return nil, ErrtreeFailedToFindNode
}
func (o *OctreeNode) Insert(position protometry.VectorN, data []interface{}) error {
	// Find the proper direction to insert
	branch := o.getOctant(position)
	// There is already a leaf there
	if o.children[branch].position != nil {
		eq, err := o.children[branch].position.ApproxEqual(position)
		if err != nil {
			return err
		}
		// Two point on same position
		if eq {
			o.children[branch].data = append(o.children[branch].data, data...)
			return nil
		}
	}

	if o.children[branch].position == nil {
		// If region node, insert in a child
		return o.children[branch].Insert(position, data)
	} else if o.children[branch].position.Dimensions[0] == math.MaxFloat64 {
		// If empty node, create node with new data on this leaf
		o.children[branch] = NewPointOctreeNode(&position, data)
		return nil
	} else {
		// If point node, store its data, make it region node,
		// move stored data down to children
		// insert new data in children
		p := *o.children[branch].position
		d := o.children[branch].data
		// Make it region node
		o.children[branch] = o.getNewRegion(branch)
		// Find new leaf for old node
		oldBranch := o.getOctant(p)
		o.children[oldBranch].Insert(p, d)
		// Find leaf for new node
		return o.children[branch].Insert(position, data)
	}
}

func (o *OctreeNode) Remove(position protometry.VectorN) error {
	n, err := o.Search(position)
	if err != nil {
		return err
	}
	if n != nil {
		*n = *NewEmptyOctreeNode()
	}
	return nil
}

const (
	TLF = iota // top left front
	TRF        // top right front
	BRF        // bottom right front
	BLF        // bottom left front
	TLB        // top left back
	TRB        // top right back
	BRB        // bottom right back
	BLB        // bottom left back
)

func (o *OctreeNode) getNewRegion(branch int) *OctreeNode {
	minD := o.region.GetMin().Dimensions
	maxD := o.region.GetMax().Dimensions
	var newNode *OctreeNode
	switch branch {
	case TLF:
		newNode = NewRegionOctreeNode(protometry.NewBox(
			*protometry.NewVectorN(
				minD[0],
				minD[1],
				minD[2]),
			*protometry.NewVectorN(
				maxD[0]/2,
				maxD[1]/2,
				maxD[2]/2)))
	case TRF:
		newNode = NewRegionOctreeNode(protometry.NewBox(
			*protometry.NewVectorN(
				maxD[0]/2,
				minD[1],
				minD[2]),
			*protometry.NewVectorN(
				maxD[0],
				maxD[1]/2,
				maxD[2]/2)))
	case BRF:
		newNode = NewRegionOctreeNode(protometry.NewBox(
			*protometry.NewVectorN(
				maxD[0]/2,
				maxD[1]/2,
				minD[2]),
			*protometry.NewVectorN(
				maxD[0],
				maxD[1],
				maxD[2]/2)))
	case BLF:
		newNode = NewRegionOctreeNode(protometry.NewBox(
			*protometry.NewVectorN(
				minD[0],
				maxD[1]/2,
				minD[2]),
			*protometry.NewVectorN(
				maxD[0]/2,
				maxD[1],
				maxD[2]/2)))
	case TLB:
		newNode = NewRegionOctreeNode(protometry.NewBox(
			*protometry.NewVectorN(
				minD[0],
				minD[1],
				maxD[2]/2),
			*protometry.NewVectorN(
				maxD[0]/2,
				maxD[1]/2,
				maxD[2])))
	case TRB:
		newNode = NewRegionOctreeNode(protometry.NewBox(
			*protometry.NewVectorN(
				maxD[0]/2,
				minD[1],
				maxD[2]/2),
			*protometry.NewVectorN(
				maxD[0],
				maxD[1]/2,
				maxD[2])))
	case BRB:
		newNode = NewRegionOctreeNode(protometry.NewBox(
			*protometry.NewVectorN(
				maxD[0]/2,
				maxD[1]/2,
				maxD[2]/2),
			*protometry.NewVectorN(
				maxD[0],
				maxD[1],
				maxD[2])))
	case BLB:
		newNode = NewRegionOctreeNode(protometry.NewBox(
			*protometry.NewVectorN(
				minD[0],
				maxD[1]/2,
				maxD[2]/2),
			*protometry.NewVectorN(
				maxD[0]/2,
				maxD[1],
				maxD[2])))
	}
	return newNode
}

func (o *OctreeNode) getOctant(position protometry.VectorN) int {
	oct := 0 // Not sure this func is correct
	center := o.region.GetCenter()
	if position.Dimensions[0] > center.Dimensions[0] {
		oct |= 4
	}
	if position.Dimensions[1] > center.Dimensions[1] {
		oct |= 2
	}
	if position.Dimensions[2] > center.Dimensions[2] {
		oct |= 1
	}
	return oct
}

// ToString Get a human readable representation of the state of
// this node and its contents.
func (o *OctreeNode) ToString() string {
	return o.recursiveToString("", "  ")
}

func (o *OctreeNode) recursiveToString(curIndent, stepIndent string) string {
	singleIndent := curIndent + stepIndent

	// default values
	childStr := "nil"
	positionStr := "nil"
	dataStr := "nil"

	if o.children[0] != nil {
		doubleIndent := singleIndent + stepIndent

		// accumulate child strings
		childStr = ""
		for i, child := range o.children {
			childStr = childStr + fmt.Sprintf("%v%d: %v,\n", doubleIndent, i, child.recursiveToString(doubleIndent, stepIndent))
		}

		childStr = fmt.Sprintf("[\n%v%v]", childStr, singleIndent)
	}

	if o.position != nil {
		positionStr = o.position.ToString()
	}

	if o.data != nil {
		// not stringifying elements since their type is unknown
		dataStr = fmt.Sprintf("[%d]", len(o.data))
	}

	return fmt.Sprintf("Node{\n%vposition: %v,\n%vdata: %v,\n%vregion: %v,\n%vchildren: %v,%v\n%v}", singleIndent, positionStr, singleIndent, dataStr, singleIndent, o.region, singleIndent, childStr, singleIndent, curIndent)
}
