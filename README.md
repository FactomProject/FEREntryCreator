# FEREntryCreator
This is a simple cli application to produce entries into the Factoid Exchange Rate (FER) chain.

The inputs:

 - The desired price change
 - The desired block height to change the price
 - The priority of this entry.  Higher priority entries trump lower priority entries.
 - The expiration block height of the entry to control entries that get lost in the network

 The outputs:

 - A cli curl command that will accomplish the entry commit
 - A cli curl command that will accomplish the reveal

