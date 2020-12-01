package retro_log

import (
	"testing"
	"time"
)

func TestVariable(t *testing.T) {
	rl := NewRetroLog("test", 1, "", 100, true)
	rl.StartTx().AppendVarStr("test", "test").Commit()
}

func TestSet(t *testing.T) {
	rl := NewRetroLog("test", 1, "", 100, true)
	rl.StartTx().AppendSetStr("testset", "123").Commit()
	rl.StartTx().AppendSetInt("testset", 124).Commit()
}

func TestExpireSet(t *testing.T) {
	rl := NewRetroLog("test", 1, "", 500, true)
	rl.CreateTimerSet("testset", 500)
	rl.StartTx().AppendSetInt("testset", 123).Commit()
	time.Sleep(100 * time.Millisecond)
	rl.StartTx().AppendSetInt("testset", 124).Commit()
	time.Sleep(3 * time.Millisecond)
	rl.StartTx().AppendVarStr("testvar", "test").Commit()
	time.Sleep(410 * time.Millisecond)
	rl.StartTx().AppendSetInt("testset", 125).Commit()
	time.Sleep(3 * time.Millisecond)
	rl.StartTx().AppendVarStr("testvar", "test2").Commit()
	time.Sleep(325 * time.Millisecond)
	rl.StartTx().AppendVarStr("testvar", "test3").Commit()
}

func TestExpireSet2(t *testing.T) {
	rl := NewRetroLog("test", 1, "", 100, true)
	rl.CreateTimerSet("testset", 3)
	rl.StartTx().AppendSetInt("testset", 123).Commit()
	rl.StartTx().AppendSetInt("testset", 124).Commit()
	time.Sleep(1 * time.Millisecond)
	rl.StartTx().AppendSetInt("testset", 125).Commit()
	time.Sleep(2 * time.Millisecond)
	rl.StartTx().AppendSetInt("testset", 126).Commit()
	rl.StartTx().AppendSetInt("testset", 127).Commit()
	rl.StartTx().AppendSetInt("testset", 128).Commit()
	time.Sleep(1 * time.Millisecond)
	rl.StartTx().AppendSetInt("testset", 129).Commit()
}


func TestExpireSet3(t *testing.T) {
	rl := NewRetroLog("test", 1, "", 100, false)
	rl.CreateTimerSet("testset", 3)
	rl.StartTx().AppendSetInt("testset", 123).Commit()
	rl.StartTx().AppendSetInt("testset", 124).Commit()
	time.Sleep(2 * time.Millisecond)
	rl.StartTx().AppendSetInt("testset", 125).Commit()
	rl.StartTx().AppendSetInt("testset", 126).Commit()
	time.Sleep(10 * time.Millisecond)
	rl.StartTx().AppendSetInt("testset", 127).Commit()
	rl.StartTx().AppendSetInt("testset", 128).Commit()
	time.Sleep(1 * time.Millisecond)
	rl.StartTx().AppendSetInt("testset", 129).Commit()
}

func TestStructInSet(t *testing.T) {
	rl := NewRetroLog("test", 1, "", 500, true)
	rqlstruct := NewRqlStruct(nil).AddVarInt("testvar", 1).AddVarStr("teststr", "teststrval")
	rl.StartTx().AppendSetStruct("testset", rqlstruct).Commit()

}

