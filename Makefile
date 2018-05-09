build:prepare
	cd build; go build ../Main/guery.go

run:stop build
	cd build; ./guery master --address 127.0.0.1:1111 >& m.log &
	cd build; \
	for i in `seq 1 30`;do \
		./guery executor --master 127.0.0.1:1111 >e$i.log &	\
	done

stop:
	-killall guery

prepare:clean
	mkdir build
	cp -rvf Master/UI ./build/

clean:stop
	-rm -rvf ./build
