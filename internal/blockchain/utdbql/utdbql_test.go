// Copyright © 2021 Kaleido, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utdbql

import (
	"context"
	"testing"

	"github.com/kaleido-io/firefly/internal/blockchain"
	"github.com/kaleido-io/firefly/internal/config"
	"github.com/kaleido-io/firefly/internal/fftypes"
	"github.com/kaleido-io/firefly/mocks/blockchainmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInit(t *testing.T) {
	u := &UTDBQL{}

	conf := config.NewPluginConfig("utdbql_unit_test")
	AddUTDBQLConf(conf)
	conf.Set(UTDBQLConfURL, "memory://")
	defer config.Reset()

	err := u.Init(context.Background(), conf, &blockchainmocks.Events{})
	assert.NoError(t, err)
	assert.NotNil(t, u.Capabilities())
	u.Close()
}

func TestInitBadURL(t *testing.T) {
	u := &UTDBQL{}

	conf := config.NewPluginConfig("utdbql_unit_test")
	AddUTDBQLConf(conf)
	conf.Set(UTDBQLConfURL, "badness://")
	defer config.Reset()

	err := u.Init(context.Background(), conf, &blockchainmocks.Events{})
	assert.Error(t, err)
}

func TestVerifyIdentitySyntaxOK(t *testing.T) {
	u := &UTDBQL{}
	id, err := u.VerifyIdentitySyntax(context.Background(), "good")
	assert.NoError(t, err)
	assert.Equal(t, "good", id)
}

func TestVerifyIdentitySyntaxFail(t *testing.T) {
	u := &UTDBQL{}
	_, err := u.VerifyIdentitySyntax(context.Background(), "!bad")
	assert.Regexp(t, "FF10131", err.Error())
}

func TestVerifyBroadcastBatchTXCycle(t *testing.T) {
	u := &UTDBQL{}
	me := &blockchainmocks.Events{}

	sbbEv := make(chan bool, 1)
	sbb := me.On("SequencedBroadcastBatch", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sbb.RunFn = func(a mock.Arguments) {
		sbbEv <- true
	}

	txEv := make(chan bool, 1)
	tx := me.On("TransactionUpdate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	tx.RunFn = func(a mock.Arguments) {
		txEv <- true
	}

	conf := config.NewPluginConfig("utdbql_unit_test")
	AddUTDBQLConf(conf)
	conf.Set(UTDBQLConfURL, "memory://")
	defer config.Reset()

	err := u.Init(context.Background(), conf, me)
	assert.NoError(t, err)
	defer u.Close()

	trackingID, err := u.SubmitBroadcastBatch(context.Background(), "id1", &blockchain.BroadcastBatch{
		Timestamp:      fftypes.NowMillis(),
		BatchID:        fftypes.NewUUID(),
		BatchPaylodRef: fftypes.NewRandB32(),
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, trackingID)

	if err == nil {
		<-txEv
		<-sbbEv
	}

}

func TestVerifyBroadcastDBError(t *testing.T) {
	u := &UTDBQL{}
	me := &blockchainmocks.Events{}

	conf := config.NewPluginConfig("utdbql_unit_test")
	AddUTDBQLConf(conf)
	conf.Set(UTDBQLConfURL, "memory://")
	defer config.Reset()

	err := u.Init(context.Background(), conf, me)
	assert.NoError(t, err)
	u.Close()

	_, err = u.SubmitBroadcastBatch(context.Background(), "id1", &blockchain.BroadcastBatch{
		Timestamp:      fftypes.NowMillis(),
		BatchID:        fftypes.NewUUID(),
		BatchPaylodRef: fftypes.NewRandB32(),
	})
	assert.Error(t, err)

}

func TestVerifyEventLoopCancelledContext(t *testing.T) {
	u := &UTDBQL{}
	me := &blockchainmocks.Events{}

	conf := config.NewPluginConfig("utdbql_unit_test")
	AddUTDBQLConf(conf)
	conf.Set(UTDBQLConfURL, "memory://")
	defer config.Reset()

	err := u.Init(context.Background(), conf, me)
	assert.NoError(t, err)
	defer u.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	u.ctx = ctx
	u.eventLoop() // Just confirming it exits
}

func TestVerifyDispatchEventBadData(t *testing.T) {
	u := &UTDBQL{}
	me := &blockchainmocks.Events{}

	conf := config.NewPluginConfig("utdbql_unit_test")
	AddUTDBQLConf(conf)
	conf.Set(UTDBQLConfURL, "memory://")
	defer config.Reset()

	err := u.Init(context.Background(), conf, me)
	assert.NoError(t, err)
	defer u.Close()

	u.dispatchEvent(&utEvent{
		txType: utDBQLEventTypeBroadcastBatch,
		data:   []byte(`!json`),
	}) // Just confirming it handles it
}