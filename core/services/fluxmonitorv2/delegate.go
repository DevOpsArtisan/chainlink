package fluxmonitorv2

import (
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/chains/evm"
	"github.com/smartcontractkit/chainlink/core/chains/evm/bulletprooftxmanager"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/services/job"
	"github.com/smartcontractkit/chainlink/core/services/keystore"
	"github.com/smartcontractkit/chainlink/core/services/pipeline"
	"github.com/smartcontractkit/sqlx"
)

// Delegate represents a Flux Monitor delegate
type Delegate struct {
	db             *sqlx.DB
	ethKeyStore    keystore.Eth
	jobORM         job.ORM
	pipelineORM    pipeline.ORM
	pipelineRunner pipeline.Runner
	chainSet       evm.ChainSet
	lggr           logger.Logger
}

var _ job.Delegate = (*Delegate)(nil)

// NewDelegate constructs a new delegate
func NewDelegate(
	ethKeyStore keystore.Eth,
	jobORM job.ORM,
	pipelineORM pipeline.ORM,
	pipelineRunner pipeline.Runner,
	db *sqlx.DB,
	chainSet evm.ChainSet,
	lggr logger.Logger,
) *Delegate {
	return &Delegate{
		db,
		ethKeyStore,
		jobORM,
		pipelineORM,
		pipelineRunner,
		chainSet,
		lggr.Named("FluxMonitor"),
	}
}

// JobType implements the job.Delegate interface
func (d *Delegate) JobType() job.Type {
	return job.FluxMonitor
}

func (Delegate) AfterJobCreated(spec job.Job)  {}
func (Delegate) BeforeJobDeleted(spec job.Job) {}

// ServicesForSpec returns the flux monitor service for the job spec
func (d *Delegate) ServicesForSpec(jb job.Job) (services []job.Service, err error) {
	if jb.FluxMonitorSpec == nil {
		return nil, errors.Errorf("Delegate expects a *job.FluxMonitorSpec to be present, got %v", jb)
	}
	chain, err := d.chainSet.Get(jb.FluxMonitorSpec.EVMChainID.ToInt())
	if err != nil {
		return nil, err
	}
	strategy := bulletprooftxmanager.NewQueueingTxStrategy(jb.ExternalJobID, chain.Config().FMDefaultTransactionQueueDepth(), chain.Config().FMSimulateTransactions())

	fm, err := NewFromJobSpec(
		jb,
		d.db,
		NewORM(d.db, d.lggr, chain.Config(), chain.TxManager(), strategy),
		d.jobORM,
		d.pipelineORM,
		NewKeyStore(d.ethKeyStore),
		chain.Client(),
		chain.LogBroadcaster(),
		d.pipelineRunner,
		chain.Config(),
		d.lggr,
	)
	if err != nil {
		return nil, err
	}

	return []job.Service{fm}, nil
}
