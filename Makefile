# run "go run . dc"
# to run the program
.PHONY: run
run:
	go run . dc

.PHONY: debug
debug:
	go run . dc -d
