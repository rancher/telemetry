var flatten = require('flat');
var fs = require('fs');
var data = JSON.parse(fs.readFileSync('./data'));

let keys = {};

for ( var i = 0 ; i < data.length ; i++ )
{
  var f = flatten(data[i]);
  Object.keys(f).forEach((k) => {
    keys[k] = 1;
  });
}

let ary = Object.keys(keys);
ary.sort();
var s = '';
ary.forEach((k) => {
  s += '"' + k + '",';
});
console.log(s);

for ( var i = 0 ; i < data.length ; i++ )
{
  var f = flatten(data[i]);
  s = '';
  ary.forEach((k) => {
    s += '"' + (f[k]||'') + '",';
  })

  console.log(s);
}
