/*
Package a2s read Steam A2S server query responses:
  - [github.com/woozymasta/a2s/pks/a2s.GetInfo]
    A2S_INFO Basic information about the server;
  - [github.com/woozymasta/a2s/pks/a2s.GetPlayers]
    A2S_PLAYER Details about each player on the server;
  - [github.com/woozymasta/a2s/pks/a2s.GetRules]
    A2S_RULES The rules the server is using;
  - [github.com/woozymasta/a2s/pks/a2s.GetChallenge]
    A2S_SERVERQUERY_GETCHALLENGE Returns a challenge number for use in the player and rules query;
  - [github.com/woozymasta/a2s/pks/a2s.GetPing]
    A2A_PING Ping the server.

More details in the official Steam documentation for the protocol [Server queries]

# Usage:

	client, err := a2s.New("127.0.0.1", 27016)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	client.SetBufferSize(2048)
	client.SetDeadlineTimeout(3)

	info, err := client.GetInfo()
	if err != nil {
		panic(err)
	}

	rules, err := client.GetRules()
	if err != nil {
		panic(err)
	}

[Server queries]: https://developer.valvesoftware.com/wiki/Server_queries
*/
package a2s
