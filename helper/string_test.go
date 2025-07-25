package helper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToInt32(t *testing.T) {
	str1 := ToString(2122003)
	v, ok := ToInt(str1)
	fmt.Println(v, ok)
}

func TestToString(t *testing.T) {
	var valueString = "bbb"
	fmt.Println(ToString(valueString))

	var valueInt = 333
	fmt.Println(ToString(valueInt))

	var valueInt32 int32 = 333
	fmt.Println(ToString(valueInt32))

	var valueInt64 int64 = 333
	fmt.Println(ToString(valueInt64))

	var valueUint uint = 333
	fmt.Println(ToString(valueUint))

	var valueUint32 uint32 = 333
	fmt.Println(ToString(valueUint32))

	var valueUint64 uint64 = 333
	fmt.Println(ToString(valueUint64))

	var valueUint8 uint8 = 10
	fmt.Println(ToString(valueUint8))
}

func TestToInt64(t *testing.T) {
	str := "actorPlayer.1"
	v, e := ToInt64(str)
	fmt.Println(v, e)
}

func TestKebabToCamel(t *testing.T) {
	assert.Equal(t, "actorPlayer", KebabToCamel("actor-player"))
	assert.Equal(t, "actorPlayer1", KebabToCamel("actor-player-1"))
	assert.Equal(t, "actorPlayer", KebabToCamel("Actor-Player"))
}

func TestPascalToCamel(t *testing.T) {
	assert.Equal(t, "actorPlayer", PascalToCamel("ActorPlayer"))
	assert.Equal(t, "actorPlayer1", PascalToCamel("ActorPlayer1"))
	assert.Equal(t, "actorPlayer-1", PascalToCamel("ActorPlayer-1"))
}

func TestPascalToSnake(t *testing.T) {
	assert.Equal(t, "actor_player", PascalToSnake("ActorPlayer"))
	assert.Equal(t, "actor_player1", PascalToSnake("ActorPlayer1"))
	assert.Equal(t, "actor_player-1", PascalToSnake("ActorPlayer-1"))
}

func TestSnakeToPascal(t *testing.T) {
	assert.Equal(t, "ActorPlayer", SnakeToPascal("actor_player"))
	assert.Equal(t, "ActorPlayer1", SnakeToPascal("actor_player1"))
	assert.Equal(t, "ActorPlayer-1", SnakeToPascal("actor_player-1"))
}
