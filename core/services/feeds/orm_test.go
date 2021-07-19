package feeds_test

import (
	"context"
	"testing"

	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"github.com/smartcontractkit/chainlink/core/internal/testutils/pgtest"
	"github.com/smartcontractkit/chainlink/core/services/feeds"
	"github.com/smartcontractkit/chainlink/core/utils/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	uri       = "http://192.168.0.1"
	name      = "Chainlink FMS"
	publicKey = crypto.PublicKey([]byte("11111111111111111111111111111111"))
	jobTypes  = pq.StringArray{feeds.JobTypeFluxMonitor, feeds.JobTypeOffchainReporting}
	network   = "mainnet"
)

func setupORM(t *testing.T) feeds.ORM {
	t.Helper()

	db := pgtest.NewGormDB(t)
	orm := feeds.NewORM(db)

	return orm
}

func Test_ORM_CreateManager(t *testing.T) {
	t.Parallel()

	orm := setupORM(t)
	mgr := &feeds.FeedsManager{
		URI:                uri,
		Name:               name,
		PublicKey:          publicKey,
		JobTypes:           jobTypes,
		Network:            network,
		IsOCRBootstrapPeer: true,
	}

	count, err := orm.CountManagers()
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	id, err := orm.CreateManager(context.Background(), mgr)
	require.NoError(t, err)

	count, err = orm.CountManagers()
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	assert.NotZero(t, id)
}

func Test_ORM_ListManagers(t *testing.T) {
	t.Parallel()

	orm := setupORM(t)
	mgr := &feeds.FeedsManager{
		URI:                uri,
		Name:               name,
		PublicKey:          publicKey,
		JobTypes:           jobTypes,
		Network:            network,
		IsOCRBootstrapPeer: true,
	}

	id, err := orm.CreateManager(context.Background(), mgr)
	require.NoError(t, err)

	mgrs, err := orm.ListManagers(context.Background())
	require.NoError(t, err)
	require.Len(t, mgrs, 1)

	actual := mgrs[0]
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, uri, actual.URI)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, publicKey, actual.PublicKey)
	assert.Equal(t, jobTypes, actual.JobTypes)
	assert.Equal(t, network, actual.Network)
	assert.Equal(t, true, actual.IsOCRBootstrapPeer)
}

func Test_ORM_GetManager(t *testing.T) {
	t.Parallel()

	orm := setupORM(t)
	mgr := &feeds.FeedsManager{
		URI:                uri,
		Name:               name,
		PublicKey:          publicKey,
		JobTypes:           jobTypes,
		Network:            network,
		IsOCRBootstrapPeer: true,
	}

	id, err := orm.CreateManager(context.Background(), mgr)
	require.NoError(t, err)

	actual, err := orm.GetManager(context.Background(), id)
	require.NoError(t, err)

	assert.Equal(t, id, actual.ID)
	assert.Equal(t, uri, actual.URI)
	assert.Equal(t, name, actual.Name)
	assert.Equal(t, publicKey, actual.PublicKey)
	assert.Equal(t, jobTypes, actual.JobTypes)
	assert.Equal(t, network, actual.Network)
	assert.Equal(t, true, actual.IsOCRBootstrapPeer)

	actual, err = orm.GetManager(context.Background(), -1)
	require.Nil(t, actual)
	require.Error(t, err)
}

func Test_ORM_CreateJobProposal(t *testing.T) {
	t.Parallel()

	orm := setupORM(t)
	fmID := createFeedsManager(t, orm)

	jp := &feeds.JobProposal{
		RemoteUUID:     uuid.NewV4(),
		Spec:           "",
		Status:         feeds.JobProposalStatusPending,
		FeedsManagerID: fmID,
	}

	count, err := orm.CountJobProposals()
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	id, err := orm.CreateJobProposal(context.Background(), jp)
	require.NoError(t, err)

	count, err = orm.CountJobProposals()
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	assert.NotZero(t, id)
}

func Test_ORM_ListJobProposals(t *testing.T) {
	t.Parallel()

	orm := setupORM(t)
	fmID := createFeedsManager(t, orm)
	uuid := uuid.NewV4()

	jp := &feeds.JobProposal{
		RemoteUUID:     uuid,
		Spec:           "",
		Status:         feeds.JobProposalStatusPending,
		FeedsManagerID: fmID,
	}

	id, err := orm.CreateJobProposal(context.Background(), jp)
	require.NoError(t, err)

	jps, err := orm.ListJobProposals(context.Background())
	require.NoError(t, err)
	require.Len(t, jps, 1)

	actual := jps[0]
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, uuid, actual.RemoteUUID)
	assert.Equal(t, jp.Status, actual.Status)
	assert.False(t, actual.JobID.Valid)
	assert.Equal(t, jp.FeedsManagerID, actual.FeedsManagerID)
}

func Test_ORM_GetJobProposal(t *testing.T) {
	t.Parallel()

	orm := setupORM(t)
	fmID := createFeedsManager(t, orm)
	uuid := uuid.NewV4()

	jp := &feeds.JobProposal{
		RemoteUUID:     uuid,
		Spec:           "",
		Status:         feeds.JobProposalStatusPending,
		FeedsManagerID: fmID,
	}

	id, err := orm.CreateJobProposal(context.Background(), jp)
	require.NoError(t, err)

	actual, err := orm.GetJobProposal(context.Background(), id)
	require.NoError(t, err)

	assert.Equal(t, id, actual.ID)
	assert.Equal(t, uuid, actual.RemoteUUID)
	assert.Equal(t, jp.Status, actual.Status)
	assert.False(t, actual.JobID.Valid)
	assert.Equal(t, jp.FeedsManagerID, actual.FeedsManagerID)

	actual, err = orm.GetJobProposal(context.Background(), int64(0))
	require.Nil(t, actual)
	require.Error(t, err)
}

func Test_ORM_UpdateJobProposalStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	orm := setupORM(t)
	fmID := createFeedsManager(t, orm)

	jp := &feeds.JobProposal{
		RemoteUUID:     uuid.NewV4(),
		Spec:           "",
		Status:         feeds.JobProposalStatusPending,
		FeedsManagerID: fmID,
	}

	id, err := orm.CreateJobProposal(ctx, jp)
	require.NoError(t, err)

	err = orm.UpdateJobProposalStatus(ctx, id, feeds.JobProposalStatusRejected)
	require.NoError(t, err)

	actual, err := orm.GetJobProposal(context.Background(), id)
	require.NoError(t, err)

	assert.Equal(t, id, actual.ID)
	assert.Equal(t, feeds.JobProposalStatusRejected, actual.Status)
}

// createFeedsManager is a test helper to create a feeds manager
func createFeedsManager(t *testing.T, orm feeds.ORM) int64 {
	mgr := &feeds.FeedsManager{
		URI:       uri,
		Name:      name,
		PublicKey: publicKey,
		JobTypes:  jobTypes,
		Network:   network,
	}

	id, err := orm.CreateManager(context.Background(), mgr)
	require.NoError(t, err)

	return id
}