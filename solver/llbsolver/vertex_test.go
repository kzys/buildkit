package llbsolver

import (
	"context"
	"testing"

	"github.com/moby/buildkit/solver/pb"
	digest "github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRecomputeDigests(t *testing.T) {
	op1 := &pb.Op{
		Op: &pb.Op_Source{
			Source: &pb.SourceOp{
				Identifier: "docker-image://docker.io/library/busybox:latest",
			},
		},
	}
	oldData, err := proto.Marshal(op1)
	require.NoError(t, err)
	oldDigest := digest.FromBytes(oldData)

	op1.GetOp().(*pb.Op_Source).Source.Identifier = "docker-image://docker.io/library/busybox:1.31.1"
	newData, err := proto.Marshal(op1)
	require.NoError(t, err)
	newDigest := digest.FromBytes(newData)

	op2 := &pb.Op{
		Inputs: []*pb.Input{
			{Digest: oldDigest.String()}, // Input is the old digest, this should be updated after recomputeDigests
		},
	}
	op2Data, err := proto.Marshal(op2)
	require.NoError(t, err)
	op2Digest := digest.FromBytes(op2Data)

	all := map[digest.Digest]*pb.Op{
		newDigest: op1,
		op2Digest: op2,
	}
	visited := map[digest.Digest]digest.Digest{oldDigest: newDigest}

	updated, err := recomputeDigests(context.Background(), all, visited, op2Digest)
	require.NoError(t, err)
	require.Len(t, visited, 2)
	require.Len(t, all, 2)
	assert.Equal(t, op1, all[newDigest])
	require.Equal(t, newDigest, visited[oldDigest])
	require.Equal(t, op1, all[newDigest])
	assert.Equal(t, op2, all[updated])
	require.Equal(t, newDigest, digest.Digest(op2.Inputs[0].Digest))
	assert.NotEqual(t, op2Digest, updated)
}
