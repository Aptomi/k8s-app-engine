Items to complete:

1. Should we have default "kind" in registry?

1. Do we need created at in metadata?

1. Add negative tests to Registry

1. Cache current Revision in Registry (for ultra fast queries to it)

1. Rename language to lang

1. Extract all (Un)marshallX from Registry to smart "Codec"s.
    it'll require passing kinds to codecs
   
1. support https://golang.org/pkg/encoding/gob/

1. re-create codec instance each time new object registered (?)
    add register to the codec interface
    
1. make following packages:
    object
    store gets codec and using it to store in db
    codec
    registry creates store and uses codecs
