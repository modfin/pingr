[@react.component]
let make = (~title, ~children=?) => {
  <div className="relative bg-gray-400 px-6 py-2">
    <p className="text-xl font-bold"> {title |> React.string} </p>
    {switch (children) {
     | Some(c) => c
     | None => React.null
     }}
  </div>;
};