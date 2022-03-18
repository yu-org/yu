go build -v -o poa

node1_path=~/run-yu/node1
node2_path=~/run-yu/node2
node3_path=~/run-yu/node3

mkdir -p $node1_path
mkdir -p $node2_path
mkdir -p $node3_path

cp poa $node1_path/
cp poa $node2_path/
mv poa $node3_path/

rm -f $node1_path/*.db
rm -f $node2_path/*.db
rm -f $node3_path/*.db
