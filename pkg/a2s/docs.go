/*
Package a2s reads Steam A2S server query responses:
  - [github.com/woozymasta/a2s/pkg/a2s.GetInfo]
    A2S_INFO Basic information about the server;
  - [github.com/woozymasta/a2s/pkg/a2s.GetPlayers]
    A2S_PLAYER Details about each player on the server;
  - [github.com/woozymasta/a2s/pkg/a2s.GetRules]
    A2S_RULES The rules the server is using;
  - [github.com/woozymasta/a2s/pkg/a2s.GetChallenge]
    A2S_SERVERQUERY_GETCHALLENGE Returns a challenge number
    for use in the player and rules query;
  - [github.com/woozymasta/a2s/pkg/a2s.GetPing]
    A2A_PING Ping the server.

More details in the official Steam documentation for the protocol [Server queries]

# Usage:

	client, err := New("127.0.0.1", 27016,
		WithBufferSize(2048),
		WithTimeout(3*time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	ctx := context.Background()

	info, err := client.GetInfo(ctx)
	if err != nil {
		panic(err)
	}

	rules, err := client.GetRules(ctx)
	if err != nil {
		panic(err)
	}

[Server queries]: https://developer.valvesoftware.com/wiki/Server_queries
*/
package a2s
