package darts

import (
	"fmt"
	"unicode/utf8"
	"sort"
)

type Node struct {
	code int
	depth int
	left int
	right int
}

func (n *Node) String() string {
	return fmt.Sprintf("node code %d depth %d left %d right %d", n.code, n.depth, n.left, n.right)
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
	if index < 0 {
		return nil
	}
	return l.nodes[index]
}

func (l *ListNode) add(node *Node) {
	l.nodes = append(l.nodes, node)
	l.size_++
}

type Word struct {
	word   string
	runes  []rune
}

func NewWord(word string) *Word {
	var runes []rune
	_word := []byte(word)
	offset := 0
	for len(_word[offset:]) > 0 {
		r, size := utf8.DecodeRune(_word[offset:])
		if r == utf8.RuneError {
			panic("invalid word")
		}
		offset += size
		runes = append(runes, r)
	}
	return &Word{word: word, runes: runes}
}

func (w *Word)GetWord() string {
	return w.word
}

func (w *Word)GetRune(index int) rune {
	if index < 0 || index >= len(w.runes) {
		panic("invalid index")
	}
	return w.runes[index]
}

func (w *Word)GetRunes() []rune {
	return w.runes
}

func (w *Word) Size() int {
	return len(w.runes)
}

func (w *Word)String() string {
	return w.word
}

type WordCodeDict struct {
	nextCode int
	dict     map[rune]int
}

func NewWordCodeDict(words []*Word) *WordCodeDict {
	var _words []rune
	dict := make(map[rune]int)
	for _, word := range words {
		for _, r := range word.GetRunes() {
			dict[r] = 0
		}
	}
	for r, _ := range dict {
		_words = append(_words, r)
	}
	sort.Sort(ByRune(_words))
	nextCode := 1
	for _, r := range _words {
		dict[r] = nextCode
		nextCode++
	}
	return &WordCodeDict{nextCode: nextCode, dict: dict}
}

func (d *WordCodeDict) Code(word rune) int {
	if code, ok := d.dict[word]; ok {
		return code
	}
	// 返回一个非法值
	return d.nextCode
}

type DoubleArrayTrie struct {
	base   []int
	check  []int
	used   []bool
	size   int
	allocSize int
	key    []*Word
	keySize int
	length []int
	value  []int
	progress int
	nextCheckPos int
	error_  int
	wordCodeDict *WordCodeDict
}

func (dat *DoubleArrayTrie)resize(newSize int) int {
	fmt.Println("resize ", newSize)
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
		// 非法单词过滤
		if dat.length != nil {
			if dat.length[i] != 0 {
				continue
			}
		} else {
			// 子节点的长度必须大于父节点的深度(即单词长度)
			// 如果len(dat.key[i]) < parent.depth,说明已经是叶子节点了
			if dat.key[i].Size() < parent.depth {
				continue
			}
		}

		tmp := dat.key[i]
		cur := 0
		if dat.length != nil {
			if dat.length[i] != 0 {
				cur = dat.wordCodeDict.Code(tmp.GetRune(parent.depth)) + 1
			}
		} else {
			if tmp.Size() != parent.depth {
				cur = dat.wordCodeDict.Code(tmp.GetRune(parent.depth)) + 1
			}
		}
		// key 必须是字典序
		if prev > cur {
			dat.error_ = -3
			return 0
		}

		// 相同前缀的节点对于父节点视为一个子节点
		//   "ab", "acz", "b"
		//              ROOT(d=0,l=0,r=3)
		//           /                     \
		//          [a(d=1,l=0,r=2)]       [b(d=1,l=2,r=3)]
		//          /               \                   /
		//         [b(d=2,l=0,r=1)] [c(d=2,l=1,r=2)]  [nil(d=2,l=2,r=3)]
		//         /                    /
		//        [nil(d=3,l=0,r=1)]   [z(d=3,l=1,r=2)]
		//                             /
		//                           [nil(d=4,l=1,r=2)]
		// 一个完整的单词最后一个结束节点的left,right与父节点保持一致
		if cur != prev || siblings.size() == 0 {
			tmp_node := new(Node)
			tmp_node.depth = parent.depth + 1
			tmp_node.code = cur
			// 左边界根据不同的前缀而不同
			tmp_node.left = i
			if siblings.size() != 0 {
				// 新的节点要加入,前一个右节点的边界需要调整,与新节点的左边界相同
				siblings.get(siblings.size() - 1).right = i
			}

			siblings.add(tmp_node)
		}

		prev = cur
	}
	// 父节点的子节点构建完成
	if siblings.size() != 0 {
		// 右边界与父节点相同
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
	// 此循环体的目标是找出满足base[begin + a1...an]==0, check[begin + a1...an]==0的n个空闲空间,a1...an是siblings中的n个节点
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
        // 这个位置已经被使用了
		if dat.used[begin] {
			continue
		}

		// 检查是否存在冲突
		for i := 0; i < siblings.size(); i++ {
			if dat.base[begin + siblings.get(i).code] != 0 {
				continue OUTER
			}
			if dat.check[begin + siblings.get(i).code] != 0 {
				continue OUTER
			}
		}
		// 找到一个没有冲突的位置
		break
	}

	if 1.0 * float64(nonzero_num) / float64(pos - dat.nextCheckPos + 1) >= 0.95 {
		dat.nextCheckPos = pos
	}
    // 标记位置被占用
	dat.used[begin] = true
	tmp_size := begin + siblings.get(siblings.size() - 1).code + 1
	// 更新 tire的size
	if dat.size < tmp_size {
		dat.size = tmp_size
	}

	// base[s] + c = t
	// check[t] = s
	for i := 0; i < siblings.size(); i++ {
		dat.check[begin + siblings.get(i).code] = begin
	}

	// 计算所有子节点的base
	for i := 0; i < siblings.size(); i++ {
		new_siblings := NewListNode()
		//// 一个词的终止且不为其他词的前缀，其实就是叶子节点
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
	*dat = DoubleArrayTrie{}
}

func (dat *DoubleArrayTrie) loseWeight() {
	base2 := make([]int, dat.size)
	check2 := make([]int, dat.size)

	if dat.allocSize > 0 {
		copy(base2, dat.base)
		copy(check2, dat.check)
	}
	dat.base = base2
	dat.check = check2
	dat.allocSize = dat.size
	return
}

func (dat *DoubleArrayTrie) GetSize() int {
	return dat.size
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
    var words []*Word
	for _, key := range _key {
		words = append(words, NewWord(key))
	}
	dat.key = words
	dat.length = _length
	dat.keySize = _keySize
	dat.value = _value
	dat.progress = 0
	dat.wordCodeDict = NewWordCodeDict(words)

	// 32个双字节
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
    dat.loseWeight()
	return dat.error_
}

func (dat *DoubleArrayTrie) ExactMatchSearch(key string) int {
	return dat.ExactMatchSearchAdvanced(key, 0, 0, 0)
}

func (dat *DoubleArrayTrie) ExactMatchSearchAdvanced(key string, pos int, length int, nodePos int) int {
	word := NewWord(key)
	if length <= 0 {
		length = word.Size()
	}
	if nodePos <= 0 {
		nodePos = 0
	}

	var result = -1

	keyChars := word.GetRunes()
	b := dat.base[nodePos]
	var p int
	for i := pos; i < length; i++ {
		p = b + dat.wordCodeDict.Code(keyChars[i]) + 1
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
	word := NewWord(key)
	if length <= 0 {
		length = word.Size()
	}
	if nodePos <= 0 {
		nodePos = 0
	}

	var result []int
	keyChars := word.GetRunes()
	b := dat.base[nodePos]
	var n, p int

	for i := pos; i < length; i++ {
		p = b
		n = dat.base[p]

		if b == dat.check[p] && n < 0 {
			result = append(result, (n * (-1) - 1))
		}

		p = b + dat.wordCodeDict.Code(keyChars[i]) + 1
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

type ByRune []rune
func (a ByRune) Len() int           { return len(a) }
func (a ByRune) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRune) Less(i, j int) bool { return a[i] < a[j] }


