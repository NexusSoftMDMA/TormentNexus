const cheerio = require("cheerio");
const content1 = '<invoke name="foo"><parameter name="code">for (let i=0; i<10; i++)';
const content2 = '<invoke name="foo"><parameter name="code">for (let i=0; i<10; i++)</parameter></invoke>';
const content3 = 'Some text before <invoke name="bar"><parameter name="cmd">ls -la';
const content4 = '<invoke   \n  name="baz"\n ><parameter \nname="script">echo "hello"';

const test = (content) => {
  const $ = cheerio.load(content, null, false);
  console.log("----");
  console.log("Input:", content);
  
  const invoke = $("invoke");
  console.log("Invoke name:", invoke.attr("name"));
  
  const parameters = $("parameter");
  console.log("Param count:", parameters.length);
  parameters.each((_, el) => {
    console.log("Param", $(el).attr("name"), "=", $(el).text());
  });
};

test(content1);
test(content2);
test(content3);
test(content4);
