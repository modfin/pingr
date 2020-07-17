module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadData
  | LoadSuccess(array(Models.Test.t))
  | LoadFail(string);

type testState =
  | NotAsked
  | Loading
  | Success(array(Models.Test.t))
  | Failure;

type state = Loadable.t(array(Models.Test.t));

[@react.component]
let make = () => {
  let (state, dispatch) =
    React.useReducer(
      (_state, action) =>
        switch (action) {
        | LoadData => Loadable.Loading
        | LoadSuccess(tests) => Loadable.Success(tests)
        | LoadFail(msg) => Loadable.Failed(msg)
        },
      Loadable.Loading,
    );

  React.useEffect0(() => {
    Api.fetchTestsWithCallback(result =>
      switch (result) {
      | None => dispatch(LoadFail("Not working"))
      | Some(tests) => dispatch(LoadSuccess(tests))
      }
    );
    None;
  });
  <>
    <div className="relative bg-gray-400 my-4 p-1">
      <p className="text-xl font-bold"> {"Active tests" |> React.string} </p>
      <button
        onClick={_event => Paths.goToNewTest()}
        className="m-1 bg-green-500 hover:bg-green-700 text-white py-1 px-2 rounded">
        {"New test" |> React.string}
      </button>
    </div>
    {switch (state) {
     | Loading => <div> {ReasonReact.string("Loading...")} </div>
     | Failed(msg) => <div> {ReasonReact.string(msg)} </div>
     | Success(tests) =>
       <table className="table-auto text-left">
         <thead>
           <tr>
             <th className="px-4 py-2"> {"Name" |> React.string} </th>
             <th className="px-4 py-2"> {"Type" |> React.string} </th>
             <th className="px-4 py-2"> {"Url" |> React.string} </th>
             <th className="px-4 py-2"> {"Created at" |> React.string} </th>
           </tr>
         </thead>
         <tbody>
           Models.Test.(
             {Array.map(
                test => {
                  <tr key={test.testId}>
                    <td className="border px-4 py-2">
                      <a
                        className="no-underline text-blue-500 hover:underline cursor-pointer"
                        onClick={_event => Paths.goToTest(test.testId)}>
                        {test.testName |> React.string}
                      </a>
                    </td>
                    <td className="border px-4 py-2">
                      {test.testType |> React.string}
                    </td>
                    <td className="border px-4 py-2">
                      {test.url |> React.string}
                    </td>
                    <td className="border px-4 py-2">
                      {test.createdAt
                       |> Js.Date.fromString
                       |> Js.Date.toLocaleString
                       |> React.string}
                    </td>
                  </tr>
                },
                tests,
              )
              |> React.array}
           )
         </tbody>
       </table>
     }}
  </>;
};