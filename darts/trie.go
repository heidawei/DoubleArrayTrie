package darts

import "fmt"

const (
	BUF_SIZE = 16384
	UNIT_SIZE = 8
)

type Node struct {
	code int
	depth int
	left int
	right int
}

type ListNode struct {
	size_   int
	nodes  []*Node
}

func NewListNode() *ListNode {
	return &ListNode{size_: 0}
}

func (l *ListNode) size() int {
	return l.size_
}

// TODO check index > size
func (l *ListNode) get(index int) *Node {
	return l.nodes[index]
}

func (l *ListNode) add(node *Node) {
	l.nodes = append(l.nodes, node)
	l.size_++
}

type DoubleArrayTrie struct {
	base   []int
	check  []int
	used   []bool
	size   int
	allocSize int
	key    []string
	keySize int
	length []int
	value  []int
	progress int
	nextCheckPos int
	error_  int
}

func (dat *DoubleArrayTrie)resize(newSize int) int {
	base2 := make([]int, newSize)
	check2 := make([]int, newSize)
	used2 := make([]bool, newSize)

	if dat.allocSize > 0 {
		copy(base2, dat.base)
		copy(check2, dat.check)
		copy(used2, dat.used)
	}
	dat.base = base2
	dat.check = check2
	dat.used = used2
	dat.allocSize = newSize
	return newSize
}

func (dat *DoubleArrayTrie) fetch(parent *Node, siblings *ListNode) int {
	if dat.error_ < 0 {
		return 0
	}
	prev := 0

	// if (dat.length != nil ? dat.length[i]:len(key[i]) < parent.depth)
	for i := parent.left; i < parent.right; i++ {
		if dat.length != nil {
			if dat.length[i] != 0 {
				continue
			}
		} else {
			if len(dat.key[i]) < parent.depth {
				continue
			}
		}

		tmp := dat.key[i]
		cur := 0
		if dat.length != nil {
			if dat.length[i] != 0 {
				cur = int(tmp[parent.depth]) + 1
			}
		} else {
			if len(tmp) != parent.depth {
				cur = int(tmp[parent.depth]) + 1
			}
		}
		if prev > cur {
			dat.error_ = -3
			return 0
		}

		if cur != prev || siblings.size() == 0 {
			tmp_node := new(Node)
			tmp_node.depth = parent.depth + 1
			tmp_node.code = cur
			tmp_node.left = i
			if siblings.size() != 0 {
				siblings.get(siblings.size() - 1).right = i
			}
			siblings.add(tmp_node)
		}

		prev = cur
	}
	if siblings.size() != 0 {
		siblings.get(siblings.size() - 1).right = parent.right
	}
	return siblings.size()
}

func (dat *DoubleArrayTrie) insert(siblings *ListNode) int {
	if dat.error_ < 0 {
		return 0
	}

	begin := 0
	nonzero_num := 0
	first := 0
	var pos int
	if siblings.get(0).code + 1 > dat.nextCheckPos {
		pos = siblings.get(0).code + 1
	} else {
		pos = dat.nextCheckPos
	}
	pos -= 1

	if dat.allocSize <= pos {
		dat.resize(pos + 1)
	}
	OUTER:
	for {
		pos++

		if dat.allocSize <= pos {
			dat.resize(pos+1)
		}
		if dat.check[pos] != 0 {
			nonzero_num++
			continue
		} else if first == 0 {
			dat.nextCheckPos = pos
			first = 1
		}

		begin = pos - siblings.get(0).code
		if dat.allocSize <= (begin + siblings.get(siblings.size() - 1).code) {
			// progress can be zero
			var l float64
			tmp_l := 1.0 * float64(dat.keySize) / float64(dat.progress + 1)
			if 1.05 > tmp_l {
				l = 1.05
			} else {
				l = tmp_l
			}
			dat.resize(int(float64(dat.allocSize) * l))
		}

		if dat.used[begin] {
			continue
		}

		for i := 0; i < siblings.size(); i++ {
			if dat.check[begin + siblings.get(0).code] != 0 {
				continue OUTER
			}
		}
		break
	}

	if 1.0 * float64(nonzero_num) / float64(pos - dat.nextCheckPos + 1) >= 0.95 {
		dat.nextCheckPos = pos
	}

	dat.used[begin] = true
	tmp_size := begin + siblings.get(siblings.size() - 1).code + 1
	if dat.size < tmp_size {
		dat.size = tmp_size
	}

	for i := 0; i < siblings.size(); i++ {
		dat.check[begin + siblings.get(i).code] = begin
	}

	for i := 0; i < siblings.size(); i++ {
		new_siblings := NewListNode()

		if dat.fetch(siblings.get(i), new_siblings) == 0 {
			if dat.value != nil {
				dat.base[begin+siblings.get(i).code] = dat.value[siblings.get(i).left - 1] * (-1) - 1
			} else {
				dat.base[begin+siblings.get(i).code] = siblings.get(i).left * (-1) - 1
			}

			if dat.value != nil && (dat.value[siblings.get(i).left] * (-1) - 1) >= 0 {
				dat.error_ = -2
				return 0
			}

			dat.progress++
		} else {
			h := dat.insert(new_siblings)
			dat.base[begin+siblings.get(i).code] = h
		}
	}

	return begin
}

func NewDoubleArrayTrie() *DoubleArrayTrie {
	return &DoubleArrayTrie{}
}

func (dat *DoubleArrayTrie) clear() {
	dat.check = nil
	dat.base = nil
	dat.used = nil
	dat.allocSize = 0
	dat.size = 0
}

func (dat *DoubleArrayTrie) GetUnitSize() int {
	return UNIT_SIZE
}

func (dat *DoubleArrayTrie) GetSize() int {
	return dat.size
}

func (dat *DoubleArrayTrie) GetTotalSize() int {
	return dat.size * UNIT_SIZE
}

func (dat *DoubleArrayTrie) GetNonzeroSize() int {
	result := 0
	for i := 0; i< dat.size; i++ {
		if dat.check[i] != 0 {
			result++
		}
	}
	return result
}

func (dat *DoubleArrayTrie) Build(_key []string) int {
	return dat.BuildAdvanced(_key, nil, nil, len(_key))
}

func (dat *DoubleArrayTrie) BuildAdvanced(_key []string, _length []int, _value []int, _keySize int) int {
	if _keySize > len(_key) || _key == nil {
		return 0
	}

	dat.key = _key
	dat.length = _length
	dat.keySize = _keySize
	dat.value = _value
	dat.progress = 0

	dat.resize(65536 * 32)

	dat.base[0] = 1
	dat.nextCheckPos = 0

	root_node := new(Node)
	root_node.left = 0
	root_node.right = dat.keySize
	root_node.depth = 0

	siblings := NewListNode()
	dat.fetch(root_node, siblings)
	dat.insert(siblings)

	dat.key = nil
	dat.used = nil

	return dat.error_
}

func (dat *DoubleArrayTrie) ExactMatchSearch(key string) int {
	return dat.ExactMatchSearchAdvanced(key, 0, 0, 0)
}

func (dat *DoubleArrayTrie) ExactMatchSearchAdvanced(key string, pos int, length int, nodePos int) int {
	if length <= 0 {
		length = len(key)
	}
	if nodePos <= 0 {
		nodePos = 0
	}

	var result = -1

	keyChars := []byte(key)
	b := dat.base[nodePos]
	var p int
	for i := pos; i < length; i++ {
		p = b + int(keyChars[i]) + 1
		if b == dat.check[p] {
			b = dat.base[p]
		} else {
			return result
		}
	}

	p = b
	n := dat.base[p]

	if b == dat.check[p] && n < 0 {
		result = n * (-1) - 1
	}
	return result
}

func (dat *DoubleArrayTrie) CommonPrefixSearch(key string) []int {
	return dat.CommonPrefixSearchAdvanced(key, 0, 0, 0)
}

func (dat *DoubleArrayTrie) CommonPrefixSearchAdvanced(key string, pos int, length int, nodePos int) []int {
	if length <= 0 {
		length = len(key)
	}
	if nodePos <= 0 {
		nodePos = 0
	}

	var result []int
	keyChars := []byte(key)
	b := dat.base[nodePos]
	var n, p int

	for i := pos; i < length; i++ {
		p = b
		n = dat.base[p]

		if b == dat.check[p] && n < 0 {
			result = append(result, (n * (-1) - 1))
		}

		p = b + int(keyChars[i]) + 1
		if b == dat.check[p] {
			b = dat.base[p]
		} else {
			return result
		}
	}

	p = b
	n = dat.base[p]

	if b == dat.check[p] && n < 0 {
		result = append(result, (n * (-1) - 1))
	}
	return result
}

func (dat *DoubleArrayTrie) Dump() {
	for i := 0; i < dat.size; i++ {
		fmt.Printf("i: %d", i)
		fmt.Printf(" [%d", dat.base[i])
		fmt.Printf(", %d]\n", dat.check[i])
	}
}


