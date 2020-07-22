let first = list => List.length(list) > 0 ? Some(List.hd(list)) : None;

let getError = (field, errors) =>
  List.filter(((name, _)) => name === field, errors)
  |> first
  |> (
    errors =>
      switch (errors) {
      | Some((_, msgs)) => List.hd(msgs) |> React.string
      | None => ReasonReact.null
      }
  );

let setJsonKey = (dict, keyName, value) => {
  switch (value) {
  | Form.Str(s) => Js.Dict.set(dict, keyName, Js.Json.string(s))
  | Form.Float(f) => Js.Dict.set(dict, keyName, Js.Json.number(f))
  | Form.Int(i) =>
    Js.Dict.set(dict, keyName, Js.Json.number(float_of_int(i)))
  | Form.List(l) =>
    let parsedJson = Js.Dict.empty();
    l
    |> List.iter(((key, value)) => {
         Js.Dict.set(parsedJson, key, Js.Json.string(value))
       });
    Js.Dict.set(dict, keyName, Js.Json.object_(parsedJson));
  };
};
let emptyMsg = "Field is required";
let aboveZero = "Field has to be > 0";