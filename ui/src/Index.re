// Entry point
[@bs.val] external document: Js.t({..}) = "document";

[%bs.raw {|require("./main.css")|}];

[%raw "require('es6-promise').polyfill()"];
[%raw "require('isomorphic-fetch')"];

// We're using raw DOM manipulations here, to avoid making you read
// ReasonReact when you might precisely be trying to learn it for the first
// time through the examples later.

let makeContainer = text => {
  let container = document##createElement("div");
  container##className #= "container";

  let title = document##createElement("div");
  title##className #= "containerTitle";
  title##innerText #= text;

  let content = document##createElement("div");
  content##className #= "containerContent";

  let () = container##appendChild(title);
  let () = container##appendChild(content);
  let () = document##body##appendChild(container);

  content;
};

ReactDOMRe.renderToElementWithId(<> <NavigationBar /> <Router /> </>, "root");