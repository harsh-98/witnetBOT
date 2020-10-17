
let bech32 = require('bech32')
let mysql = require('mysql')


let connection = mysql.createConnection({
  host:'localhost',
  user: '',
  password: '',
  database: ''
})
var nodeIDs=[]
connection.connect()
connection.query('select NodeID from userNodeMap', function(err, res, fields){
    // connection.release()
  console.log(err)
  res.forEach(function(e){
    let oldA =e.NodeID
  if(oldA[0] !='t') return
    try{
    var decoded =bech32.decode(oldA)
    }catch(e){return}
  let addr=bech32.encode("wit", decoded.words)
    connection.query("update userNodeMap set NodeID=? where NodeID=?", [addr, oldA], function(er, re, field){
      if(er) console.log(er);
      // connection.release()
    })
  })
})


