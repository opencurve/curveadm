package main

func main() {
	r := create()
	err := r.init()
	if err != nil {
		panic(err)
	}

	r.run()
	r.wait()
}
