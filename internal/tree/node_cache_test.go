package tree

import "testing"

func expectNodeCachePanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	fn()
}

func cachedElementForTest(
	t *testing.T,
	cache *NodeCache,
	kind RawSyntaxKind,
	value string,
) hashedGreenElement {
	t.Helper()

	hash, token := cache.token(kind, value)
	if hash == uncachedNodeHash {
		t.Fatal("interned token returned uncached hash")
	}

	return hashedGreenElement{
		hash:    hash,
		element: token,
	}
}

func TestNodeCacheInternsTokens(t *testing.T) {
	cache := NewNodeCache()
	kind := RawSyntaxKind(1)

	hash1, first := cache.token(kind, "yes")
	hash2, second := cache.token(kind, "yes")

	if first != second {
		t.Fatal("identical tokens were not interned to the same pointer")
	}
	if hash1 != hash2 {
		t.Fatalf("token hashes differ: %d != %d", hash1, hash2)
	}
	if hash1 == uncachedNodeHash {
		t.Fatal("interned token returned uncached hash")
	}
}

func TestNodeCacheDistinguishesTokenKindAndText(t *testing.T) {
	cache := NewNodeCache()

	_, identifierYes := cache.token(RawSyntaxKind(1), "yes")
	_, identifierNo := cache.token(RawSyntaxKind(1), "no")
	_, otherKindYes := cache.token(RawSyntaxKind(2), "yes")

	if identifierYes == identifierNo {
		t.Fatal("tokens with different text were interned together")
	}
	if identifierYes == otherKindYes {
		t.Fatal("tokens with different kinds were interned together")
	}
}

func TestNodeCacheInternsGenericTriviaAcrossDifferentTokens(t *testing.T) {
	cache := NewNodeCache()
	leading := []TriviaPiece{
		NewNewlineTriviaPiece(1),
		NewWhitespaceTriviaPiece(1),
	}

	_, first := cache.tokenWithTrivia(RawSyntaxKind(1), "\n a", leading, nil)
	_, second := cache.tokenWithTrivia(RawSyntaxKind(1), "\n b", leading, nil)

	if first == second {
		t.Fatal("different token text should not intern to the same token")
	}
	if first.leading.data == nil || second.leading.data == nil {
		t.Fatal("expected non-empty interned leading trivia")
	}
	if first.leading.data != second.leading.data {
		t.Fatal("identical trivia descriptions were not interned together")
	}
}

func TestNodeCacheUsesSingleWhitespaceFastPath(t *testing.T) {
	cache := NewNodeCache()
	space := []TriviaPiece{NewWhitespaceTriviaPiece(1)}

	_, first := cache.tokenWithTrivia(RawSyntaxKind(1), " a", space, nil)
	_, second := cache.tokenWithTrivia(RawSyntaxKind(1), " b", space, nil)

	if cache.trivia.whitespace.data == nil {
		t.Fatal("single-space trivia fast path was not initialized")
	}
	if first.leading.data != cache.trivia.whitespace.data {
		t.Fatal("first token did not use shared single-space trivia")
	}
	if second.leading.data != cache.trivia.whitespace.data {
		t.Fatal("second token did not use shared single-space trivia")
	}
}

func TestNodeCacheVacantNodeKeepsComputedHashAndInterns(t *testing.T) {
	cache := NewNodeCache()
	children := []hashedGreenElement{
		cachedElementForTest(t, cache, RawSyntaxKind(1), "foo"),
		cachedElementForTest(t, cache, RawSyntaxKind(2), "="),
		cachedElementForTest(t, cache, RawSyntaxKind(1), "yes"),
	}
	nodeKind := RawSyntaxKind(100)

	entry := cache.node(nodeKind, children)
	if entry.state != nodeCacheEntryVacant {
		t.Fatalf("first lookup state = %v, want vacant", entry.state)
	}
	if entry.hash == uncachedNodeHash {
		t.Fatal("vacant cacheable node lost its computed hash")
	}

	nodeChildren := make([]GreenElement, len(children))
	for i, child := range children {
		nodeChildren[i] = child.element
	}
	node := NewGreenNode(nodeKind, nodeChildren)

	cachedHash := entry.cacheNode(node)
	if cachedHash != entry.hash {
		t.Fatalf("cacheNode() hash = %d, want %d", cachedHash, entry.hash)
	}
	if cachedHash == uncachedNodeHash {
		t.Fatal("cached node returned uncached hash")
	}

	secondEntry := cache.node(nodeKind, children)
	if secondEntry.state != nodeCacheEntryCached {
		t.Fatalf("second lookup state = %v, want cached", secondEntry.state)
	}
	if secondEntry.cached != node {
		t.Fatal("cached node lookup did not reuse the original node pointer")
	}
	if secondEntry.hash != cachedHash {
		t.Fatalf("cached lookup hash = %d, want %d", secondEntry.hash, cachedHash)
	}
}

func TestNodeCacheDoesNotCacheNodeWithTooManyChildren(t *testing.T) {
	cache := NewNodeCache()
	child := cachedElementForTest(t, cache, RawSyntaxKind(1), "x")
	children := []hashedGreenElement{child, child, child, child}

	entry := cache.node(RawSyntaxKind(100), children)

	if entry.state != nodeCacheEntryNoCache {
		t.Fatalf("state = %v, want no-cache", entry.state)
	}
	if entry.hash != uncachedNodeHash {
		t.Fatalf("hash = %d, want uncached hash", entry.hash)
	}
}

func TestNodeCachePropagatesUncachedChild(t *testing.T) {
	cache := NewNodeCache()
	token := NewGreenToken(RawSyntaxKind(1), "x")
	children := []hashedGreenElement{
		{
			hash:    uncachedNodeHash,
			element: token,
		},
	}

	entry := cache.node(RawSyntaxKind(100), children)

	if entry.state != nodeCacheEntryNoCache {
		t.Fatalf("state = %v, want no-cache", entry.state)
	}
	if entry.hash != uncachedNodeHash {
		t.Fatalf("hash = %d, want uncached hash", entry.hash)
	}
}

func TestNodeCacheGenerationRetainsOnlyReusedTokens(t *testing.T) {
	cache := NewNodeCache()
	kind := RawSyntaxKind(1)

	_, stale := cache.token(kind, "stale")
	_, live := cache.token(kind, "live")
	if len(cache.tokens) != 2 {
		t.Fatalf("token cache size = %d, want 2", len(cache.tokens))
	}

	cache.incrementGeneration()
	_, liveAgain := cache.token(kind, "live")
	if liveAgain != live {
		t.Fatal("token was not reused in the next generation")
	}

	cache.retainCache()

	if len(cache.tokens) != 1 {
		t.Fatalf("token cache size after retain = %d, want 1", len(cache.tokens))
	}
	if _, ok := cache.tokens[tokenCacheKey{kind: kind, value: stale.value}]; ok {
		t.Fatal("stale token remained in cache")
	}
	if _, ok := cache.tokens[tokenCacheKey{kind: kind, value: live.value}]; !ok {
		t.Fatal("reused token was removed from cache")
	}
}

func TestTriviaCacheGenerationRetainsOnlyReusedGenericTrivia(t *testing.T) {
	var cache triviaCache
	firstPieces := []TriviaPiece{
		NewNewlineTriviaPiece(1),
		NewWhitespaceTriviaPiece(2),
	}
	stalePieces := []TriviaPiece{NewCommentTriviaPiece(4)}

	first := cache.getTrivia(cacheGenerationA, firstPieces)
	_ = cache.getTrivia(cacheGenerationA, stalePieces)

	firstAgain := cache.getTrivia(cacheGenerationB, firstPieces)
	if first.data != firstAgain.data {
		t.Fatal("generic trivia was not reused across generations")
	}

	cache.retainTrivia(cacheGenerationB)

	entryCount := 0
	for _, bucket := range cache.cache {
		entryCount += len(bucket)
	}
	if entryCount != 1 {
		t.Fatalf("generic trivia entry count after retain = %d, want 1", entryCount)
	}
}

func TestNodeCacheEntryCacheNodeRejectsInvalidUse(t *testing.T) {
	t.Run("wrong state", func(t *testing.T) {
		entry := nodeCacheEntry{state: nodeCacheEntryNoCache}
		expectNodeCachePanic(t, func() {
			entry.cacheNode(NewGreenNode(RawSyntaxKind(100), nil))
		})
	})

	t.Run("nil node", func(t *testing.T) {
		cache := NewNodeCache()
		entry := cache.node(RawSyntaxKind(100), nil)
		expectNodeCachePanic(t, func() {
			entry.cacheNode(nil)
		})
	})

	t.Run("different kind", func(t *testing.T) {
		cache := NewNodeCache()
		entry := cache.node(RawSyntaxKind(100), nil)
		expectNodeCachePanic(t, func() {
			entry.cacheNode(NewGreenNode(RawSyntaxKind(101), nil))
		})
	})
}

func TestHashHelpersReserveZeroForUncachedNodes(t *testing.T) {
	if usableHash(uncachedNodeHash) == uncachedNodeHash {
		t.Fatal("usableHash did not remap the reserved zero hash")
	}

	if tokenHash(RawSyntaxKind(1), "foo") == uncachedNodeHash {
		t.Fatal("tokenHash returned the reserved uncached hash")
	}

	pieces := []TriviaPiece{NewCommentTriviaPiece(4)}
	if triviaHash(pieces) == uncachedNodeHash {
		t.Fatal("triviaHash returned the reserved uncached hash")
	}
}
