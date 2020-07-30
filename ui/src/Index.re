// Entry point
[@bs.val] external document: Js.t({..}) = "document";

[%bs.raw {|require("../../css/main.css")|}];

[%raw "require('es6-promise').polyfill()"];
[%raw "require('isomorphic-fetch')"];

ReactDOMRe.renderToElementWithId(<> <NavigationBar /> <Router /> </>, "root");