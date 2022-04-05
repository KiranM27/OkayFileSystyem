Every thing in MetaData must be a sync map:

- Upon appendRequest, we spin up a go routine to handle this and make changes to meta data. But if there is multiple clients asking to append then we will have a problem.

- Same thing with replicate. Each time we replicate, we spin up a go routine for the dead chunk server. If there is one chunk server then there wont be an issue.
  All ports still seem to be hard coded

```
func choose_3_random_chunkServers() []int {

	chunkServerArray := map[int]bool{
		8081: false,
		8082: false,
		8083: false,
		8084: false,
		8085: false,
	}
```
