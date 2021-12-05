go build -v -o poa

cp poa ~/run-yu/node1/

cp poa ~/run-yu/node2/

mv poa ~/run-yu/node3/

rm -f ~/run-yu/node1/*.db

rm -f ~/run-yu/node2/*.db

rm -f ~/run-yu/node3/*.db
