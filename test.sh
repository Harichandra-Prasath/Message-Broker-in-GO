for ((i=0;i<50;i++)); do
    c="bar$i"
    d="foo$i"
    curl -X POST http://127.0.0.1:3000/pub/bar --data-binary $c
    curl -X POST http://127.0.0.1:3000/pub/foo --data-binary $d
done