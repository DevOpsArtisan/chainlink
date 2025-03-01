package ocrcommon_test

import (
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink/core/services/ocrcommon"

	"github.com/smartcontractkit/chainlink/core/chains/evm/bulletprooftxmanager"
	bptxmmocks "github.com/smartcontractkit/chainlink/core/chains/evm/bulletprooftxmanager/mocks"
	"github.com/smartcontractkit/chainlink/core/internal/cltest"
	"github.com/smartcontractkit/chainlink/core/internal/testutils/pgtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Transmitter_CreateEthTransaction(t *testing.T) {
	db := pgtest.NewSqlxDB(t)
	cfg := cltest.NewTestGeneralConfig(t)
	ethKeyStore := cltest.NewKeyStore(t, db, cfg).Eth()

	_, fromAddress := cltest.MustInsertRandomKey(t, ethKeyStore, 0)

	gasLimit := uint64(1000)
	toAddress := cltest.NewAddress()
	payload := []byte{1, 2, 3}
	txm := new(bptxmmocks.TxManager)
	strategy := new(bptxmmocks.TxStrategy)

	transmitter := ocrcommon.NewTransmitter(txm, fromAddress, gasLimit, strategy)

	txm.On("CreateEthTransaction", bulletprooftxmanager.NewTx{
		FromAddress:    fromAddress,
		ToAddress:      toAddress,
		EncodedPayload: payload,
		GasLimit:       gasLimit,
		Meta:           nil,
		Strategy:       strategy,
	}, mock.Anything).Return(bulletprooftxmanager.EthTx{}, nil).Once()
	require.NoError(t, transmitter.CreateEthTransaction(context.Background(), toAddress, payload))

	txm.AssertExpectations(t)
}
