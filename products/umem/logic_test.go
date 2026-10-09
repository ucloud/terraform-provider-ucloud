package umem

import (
	"testing"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
)

func TestDistributedBlockCnt(t *testing.T) {
	blocks := []umem.UMemBlockInfo{{BlockId: "block-a"}, {BlockId: "block-b"}}

	cases := []struct {
		name       string
		blocks     []umem.UMemBlockInfo
		readable   bool
		configured int
		want       int
	}{
		{name: "live shard count wins", blocks: blocks, readable: true, configured: 0, want: 2},
		{name: "live shard count wins over stale config", blocks: blocks, readable: true, configured: 4, want: 2},
		{name: "unreadable block info keeps configured", blocks: nil, readable: false, configured: 4, want: 4},
		{name: "unreadable and unset stays zero", blocks: nil, readable: false, configured: 0, want: 0},
		{name: "readable but empty keeps configured", blocks: nil, readable: true, configured: 4, want: 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := distributedBlockCnt(tc.blocks, tc.readable, tc.configured); got != tc.want {
				t.Fatalf("distributedBlockCnt = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestDistributedBlockSizeForResize(t *testing.T) {
	cases := []struct {
		name      string
		total     int
		shards    int
		want      int
		wantError bool
	}{
		{name: "two shards of eight", total: 16, shards: 2, want: 8},
		{name: "four shards of four", total: 16, shards: 4, want: 4},
		{name: "twenty GB shard", total: 80, shards: 4, want: 20},
		{name: "not divisible by shard count", total: 20, shards: 3, wantError: true},
		{name: "invalid per-shard size", total: 12, shards: 2, wantError: true},
		{name: "no shards reported", total: 16, shards: 0, wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := distributedBlockSizeForResize(tc.total, tc.shards)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected an error, got shard size %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("distributedBlockSizeForResize = %d, want %d", got, tc.want)
			}
		})
	}
}
