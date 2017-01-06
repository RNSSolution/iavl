package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/abci/server"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/merkleeyes/app"
	eyes "github.com/tendermint/merkleeyes/client"
)

var abciType = "socket"

func TestNonPersistent(t *testing.T) {
	testProcedure(t, "tcp://127.0.0.1:46659", "", 0, false, true)
}

func TestPersistent(t *testing.T) {
	testProcedure(t, "tcp://127.0.0.1:46659", "", 0, false, false)
	//testProcedure(t, "tcp://127.0.0.1:46669", "", 0, true, true)
}

func testProcedure(t *testing.T, addr, dbName string, cache int, testPersistence, clearRecords bool) {

	// Start the listener
	mApp := app.NewMerkleEyesApp(dbName, cache)
	s, err := server.NewServer(addr, abciType, mApp)

	if err != nil {
		t.Fatal(err.Error())
		return
	}
	defer s.Stop()

	// Create client
	cli, err := eyes.NewClient(addr, abciType)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	defer cli.Stop()

	if !testPersistence {
		// Empty
		commit(t, cli, "")
		get(t, cli, "foo", "", "")
		get(t, cli, "bar", "", "")
		// Set foo=FOO
		set(t, cli, "foo", "FOO")

		commit(t, cli, "F5EB586B3499DA359ECF3BD2B2DCCE8C97C5E479")
		get(t, cli, "foo", "FOO", "")
		get(t, cli, "foa", "", "")
		get(t, cli, "foz", "", "")
		rem(t, cli, "foo")
		// Empty
		get(t, cli, "foo", "", "")
		commit(t, cli, "")
		// Set foo1, foo2, foo3...
		set(t, cli, "foo1", "1")
		set(t, cli, "foo2", "2")
		set(t, cli, "foo3", "3")
		set(t, cli, "foo1", "4")
		get(t, cli, "foo1", "4", "")
		get(t, cli, "foo2", "2", "")
		get(t, cli, "foo3", "3", "")
		commit(t, cli, "FB3B1F101D5059C75455F8476A772CDFCF12B440")
	} else {
		get(t, cli, "foo1", "4", "")
		get(t, cli, "foo2", "2", "")
		get(t, cli, "foo3", "3", "")

	}

	if clearRecords {
		rem(t, cli, "foo3")
		rem(t, cli, "foo2")
		rem(t, cli, "foo1")
		// Empty
		commit(t, cli, "")
	}
}

func get(t *testing.T, cli *eyes.Client, key string, value string, err string) {
	res := cli.GetSync([]byte(key))
	val := []byte(nil)
	if value != "" {
		val = []byte(value)
	}
	require.EqualValues(t, val, res.Data)
	if res.IsOK() {
		assert.Equal(t, "", err)
	} else {
		assert.NotEqual(t, "", err)
	}
}

func set(t *testing.T, cli *eyes.Client, key string, value string) {
	cli.SetSync([]byte(key), []byte(value))
}

func rem(t *testing.T, cli *eyes.Client, key string) {
	cli.RemSync([]byte(key))
}

func commit(t *testing.T, cli *eyes.Client, hash string) {
	res := cli.CommitSync()
	require.False(t, res.IsErr(), res.Error())
	assert.Equal(t, hash, Fmt("%X", res.Data))
}
