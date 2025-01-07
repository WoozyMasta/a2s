/*
An extension for the A2S client that overrides the standard GetRules() method
to work with the Arma 3 Server Browser Protocol (A3SBP), built on top of A2S_RULES

https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol3

# Usage:

	client, err := a2s.New("127.0.0.1", 27016)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Wrap client
	a3Client := &a3sb.Client{Client: client}

	// Game id must be passed as the second argument to properly read the Arma 3 or Dayz rules
	rules, err := a3Client.GetRules(221100)
	if err != nil {
		panic(err)
	}

	// Can also perform standard a2s methods
	info, err := a3Client.GetInfo()
	if err != nil {
		panic(err)
	}
*/
package a3sb
