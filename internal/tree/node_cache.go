package tree

const (
	// optimization path: benchmark max child nodes to cache
	maxCachedNodeChildren = 3
	// un-interned green node
	uncachedNodeHash = uint64(0)
	// optimization path: benchmark fnv hashing lib
	fnv64OffsetBasis = uint64(14695981039346656037)
	fnv64Prime       = uint64(1099511628211)
)

// cacheGeneration identifies one of the two generations retained by NodeCache.
type cacheGeneration uint8

const (
	cacheGenerationA cacheGeneration = iota
	cacheGenerationB
)

func (cg cacheGeneration) next() cacheGeneration {
	if cg == cacheGenerationA {
		return cacheGenerationB
	}
	return cacheGenerationA
}

// hashedGreenElement pairs an interned green element with its precomputed hash.
type hashedGreenElement struct {
	hash    uint64
	element GreenElement
}

// tokenCacheKey identifies an interned green token.
type tokenCacheKey struct {
	kind  RawSyntaxKind
	value string
}

// cachedToken stores the token metadata needed by NodeCache.
type cachedToken struct {
	token      *GreenToken
	hash       uint64
	generation cacheGeneration
}

// nodeCacheKey identifies an interned green node.
type nodeCacheKey struct {
	kind     RawSyntaxKind
	children [maxCachedNodeChildren]GreenElement
}

// cachedNode stores the node metadata needed by NodeCache.
type cachedNode struct {
	node       *GreenNode
	hash       uint64
	generation cacheGeneration
}

// cachedTrivia stores one interned trivia value and its cache generation.
type cachedTrivia struct {
	trivia     greenTrivia
	generation cacheGeneration
}

// triviaCache interns green trivia.
type triviaCache struct {
	cache map[uint64][]cachedTrivia
	// extremely common special whitespace(1) trivia
	whitespace greenTrivia
}

// NodeCache interns immutable green trivia, tokens, and green nodes.
type NodeCache struct {
	nodes      map[nodeCacheKey]cachedNode
	tokens     map[tokenCacheKey]cachedToken
	trivia     triviaCache
	generation cacheGeneration
}

type nodeCacheEntryState uint8

const (
	// not cacheable, too many children or a child wasn't interned
	nodeCacheEntryNoCache nodeCacheEntryState = iota
	// cacheable, not in cache
	nodeCacheEntryVacant
	// already in cache
	nodeCacheEntryCached
)

// nodeCacheEntry represents the result of looking up a node.
type nodeCacheEntry struct {
	cache  *NodeCache
	state  nodeCacheEntryState
	key    nodeCacheKey
	hash   uint64
	cached *GreenNode
}

// NewNodeCache returns an empty node cache.
func NewNodeCache() *NodeCache {
	return &NodeCache{}
}

func (nc *NodeCache) init() {
	if nc.nodes == nil {
		nc.nodes = make(map[nodeCacheKey]cachedNode)
	}
	if nc.tokens == nil {
		nc.tokens = make(map[tokenCacheKey]cachedToken)
	}
	nc.trivia.initTriviaCacheWhitespace()
}

// token returns an interned token without trivia and its precomputed hash.
func (nc *NodeCache) token(kind RawSyntaxKind, value string) (uint64, *GreenToken) {
	return nc.tokenWithTrivia(kind, value, nil, nil)
}

// tokenWithTrivia returns an interned token and its precomputed hash.
func (nc *NodeCache) tokenWithTrivia(kind RawSyntaxKind, value string, leading, trailing []TriviaPiece) (uint64, *GreenToken) {
	nc.init()
	lookupKey := tokenCacheKey{kind: kind, value: value}
	if entry, ok := nc.tokens[lookupKey]; ok {
		entry.generation = nc.generation
		nc.tokens[lookupKey] = entry
		return entry.hash, entry.token
	}
	hash := tokenHash(kind, value)
	leadingTrivia := nc.trivia.getTrivia(nc.generation, leading)
	trailingTrivia := nc.trivia.getTrivia(nc.generation, trailing)
	token := newGreenTokenWithTrivia(kind, value, leadingTrivia, trailingTrivia)
	key := tokenCacheKey{kind: kind, value: token.value}
	nc.tokens[key] = cachedToken{
		token:      token,
		hash:       hash,
		generation: nc.generation,
	}
	return hash, token
}

// node looks up an interned node with the given kind and children.
func (nc *NodeCache) node(kind RawSyntaxKind, children []hashedGreenElement) nodeCacheEntry {
	nc.init()
	if len(children) > maxCachedNodeChildren {
		return nodeCacheEntry{
			cache: nc,
			state: nodeCacheEntryNoCache,
			hash:  uncachedNodeHash,
		}
	}
	key := nodeCacheKey{kind: kind}
	hash := hashKind(fnv64OffsetBasis, kind)
	for i, child := range children {
		validateGreenElement(child.element)
		if child.hash == uncachedNodeHash {
			return nodeCacheEntry{
				cache: nc,
				state: nodeCacheEntryNoCache,
				hash:  uncachedNodeHash,
			}
		}
		key.children[i] = child.element
		hash = hashUint64(hash, child.hash)
	}
	hash = usableHash(hash)
	if entry, ok := nc.nodes[key]; ok {
		entry.generation = nc.generation
		nc.nodes[key] = entry
		return nodeCacheEntry{
			cache:  nc,
			state:  nodeCacheEntryCached,
			key:    key,
			hash:   entry.hash,
			cached: entry.node,
		}
	}
	return nodeCacheEntry{
		cache: nc,
		state: nodeCacheEntryVacant,
		key:   key,
		hash:  hash,
	}
}

// cacheNode inserts node into a vacant node cache entry and returns its hash
func (e *nodeCacheEntry) cacheNode(node *GreenNode) uint64 {
	if e.state != nodeCacheEntryVacant {
		panic("tree: node cache entry is not vacant")
	}
	if node == nil {
		panic("tree: cannot cache a nil green node")
	}
	if node.Kind() != e.key.kind {
		panic("tree: cached node kind does not match lookup kind")
	}
	e.cache.nodes[e.key] = cachedNode{
		node:       node,
		hash:       e.hash,
		generation: e.cache.generation,
	}
	return e.hash
}

// incrementGeneration starts a new cache generation
func (nc *NodeCache) incrementGeneration() {
	nc.generation = nc.generation.next()
}

// retainCache removes entries that were not used in the current generation.
func (nc *NodeCache) retainCache() {
	for key, entry := range nc.nodes {
		if entry.generation != nc.generation {
			delete(nc.nodes, key)
		}
	}
	for key, entry := range nc.tokens {
		if entry.generation != nc.generation {
			delete(nc.tokens, key)
		}
	}
	nc.trivia.retainTrivia(nc.generation)
}

func (tc *triviaCache) initTriviaCacheWhitespace() {
	if tc.whitespace.data == nil {
		tc.whitespace = newGreenTrivia([]TriviaPiece{NewWhitespaceTriviaPiece(1)})
	}
}

// getTrivia returns the interned green trivia represented by pieces.
func (tc *triviaCache) getTrivia(generation cacheGeneration, pieces []TriviaPiece) greenTrivia {
	if len(pieces) == 0 {
		return greenTrivia{}
	}
	if len(pieces) == 1 && pieces[0].Kind() == TriviaWhitespace && pieces[0].TextLen() == 1 {
		tc.initTriviaCacheWhitespace()
		return tc.whitespace
	}
	if tc.cache == nil {
		tc.cache = make(map[uint64][]cachedTrivia)
	}
	hash := triviaHash(pieces)
	bucket := tc.cache[hash]
	for index := range bucket {
		entry := &bucket[index]
		if isTriviaEqual(entry.trivia, pieces) {
			entry.generation = generation
			return entry.trivia
		}
	}
	trivia := newGreenTrivia(pieces)
	tc.cache[hash] = append(bucket, cachedTrivia{
		trivia:     trivia,
		generation: generation,
	})
	return trivia
}

// retainTrivia removes trivia entries not used in current generation.
func (tc *triviaCache) retainTrivia(generation cacheGeneration) {
	for hash, bucket := range tc.cache {
		kept := bucket[:0]
		for _, entry := range bucket {
			if entry.generation == generation {
				kept = append(kept, entry)
			}
		}
		if len(kept) == 0 {
			delete(tc.cache, hash)
			continue
		}
		clear(bucket[len(kept):])
		tc.cache[hash] = kept
	}
}

func isTriviaEqual(trivia greenTrivia, pieces []TriviaPiece) bool {
	if trivia.len() != len(pieces) {
		return false
	}
	for i, piece := range pieces {
		cached := trivia.data.pieces[i]
		if cached.Kind() != piece.Kind() || cached.TextLen() != piece.TextLen() {
			return false
		}
	}
	return true
}

func tokenHash(kind RawSyntaxKind, value string) uint64 {
	hash := hashKind(fnv64OffsetBasis, kind)
	for index := 0; index < len(value); index++ {
		hash ^= uint64(value[index])
		hash *= fnv64Prime
	}
	return usableHash(hash)
}

func hashKind(hash uint64, kind RawSyntaxKind) uint64 {
	value := uint64(kind)

	hash ^= value & 0xff
	hash *= fnv64Prime

	hash ^= (value >> 8) & 0xff
	hash *= fnv64Prime

	return hash
}

func hashUint64(hash, value uint64) uint64 {
	for shift := uint(0); shift < 64; shift += 8 {
		hash ^= (value >> shift) & 0xff
		hash *= fnv64Prime
	}
	return hash
}

// usableHash ensures zero remains reserved for uncached nodes.
func usableHash(hash uint64) uint64 {
	if hash == uncachedNodeHash {
		return 1
	}
	return hash
}

// triviaHash returns a hash for a sequence of trivia pieces.
func triviaHash(pieces []TriviaPiece) uint64 {
	hash := hashUint64(fnv64OffsetBasis, uint64(len(pieces)))
	for _, piece := range pieces {
		hash ^= uint64(piece.Kind())
		hash *= fnv64Prime
		hash = hashUint64(hash, uint64(piece.TextLen()))
	}
	return usableHash(hash)
}
