go build -v -o poa

node1_path=~/run-yu/node1
node2_path=~/run-yu/node2
node3_path=~/run-yu/node3

node1_cfg_path=yu_conf/node1/kernel.toml
node2_cfg_path=yu_conf/node2/kernel.toml
node3_cfg_path=yu_conf/node3/kernel.toml

yu_cfg_path=/yu_conf

mkdir -p $node1_path/$yu_cfg_path
mkdir -p $node2_path/$yu_cfg_path
mkdir -p $node3_path/$yu_cfg_path

cp poa $node1_path/
cp $node1_cfg_path  $node1_path/$yu_cfg_path
cp poa $node2_path/
cp $node2_cfg_path  $node2_path/$yu_cfg_path
mv poa $node3_path/
cp $node3_cfg_path  $node3_path/$yu_cfg_path

rm -rf $node1_path/yu
rm -rf $node2_path/yu
rm -rf $node3_path/yu
