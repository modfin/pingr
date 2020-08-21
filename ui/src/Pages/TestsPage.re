module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadData
  | LoadSuccess(array(Models.TestStatus.t))
  | LoadFail(string);

type testState =
  | NotAsked
  | Loading
  | Success(array(Models.TestStatus.t))
  | Failure;

type state = Loadable.t(array(Models.TestStatus.t));

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

  React.useEffect1(
    () => {
      if (state == Loadable.Loading) {
        Api.fetchTestsStatusWithCallback(result =>
          switch (result) {
          | None =>
            dispatch(
              LoadFail("Not working, perhaps you haven't added any tests?"),
            )
          | Some(tests) => dispatch(LoadSuccess(tests))
          }
        );
      };
      None;
    },
    [|state|],
  );
  <>
    <Divider title="Active tests">
      <button
        onClick={_event => Paths.goToNewTest()}
        className="my-1 bg-green-500 hover:bg-green-700 text-white py-1 px-2 rounded">
        {"New test" |> React.string}
      </button>
      <button
        onClick={_event => dispatch(LoadData)}
        className="my-1 ml-2 bg-blue-500 hover:bg-blue-700 text-white py-1 px-2 rounded">
        {"Refresh" |> React.string}
      </button>
    </Divider>
    <div className="px-6 pb-6">
      {switch (state) {
       | Loading =>
         <div className="py-2"> {ReasonReact.string("Loading...")} </div>
       | Failed(msg) =>
         <div className="py-2"> {ReasonReact.string(msg)} </div>
       | Success(testsStatus) =>
         <table className="table-auto text-left">
           <thead>
             <tr>
               <th className="px-4 py-2"> {"" |> React.string} </th>
               <th className="px-4 py-2"> {"Name" |> React.string} </th>
               <th className="px-4 py-2"> {"Type" |> React.string} </th>
               <th className="px-4 py-2"> {"Url" |> React.string} </th>
               <th className="px-4 py-2">
                 {"Response time (ms)" |> React.string}
               </th>
             </tr>
           </thead>
           <tbody>
             Models.TestStatus.(
               {Array.map(
                  testStatus => {
                    <tr key={testStatus.testId}>
                      <td className="border px-2 py-2">
                        <TestStatusLabel
                          active={testStatus.active}
                          statusId={testStatus.statusId}
                        />
                      </td>
                      <td className="border px-4 py-2">
                        <a
                          className="no-underline text-blue-500 hover:underline cursor-pointer"
                          onClick={_event =>
                            Paths.goToTest(testStatus.testId)
                          }>
                          {testStatus.testName |> React.string}
                        </a>
                      </td>
                      <td className="border px-4 py-2">
                        {testStatus.testType |> React.string}
                      </td>
                      <td className="border px-4 py-2">
                        {(
                           switch (testStatus.testType) {
                           | "HTTPPush"
                           | "PrometheusPush" =>
                             "/api/push/"
                             ++ testStatus.testId
                             ++ "/"
                             ++ testStatus.testName
                             |> String.map(c =>
                                  if (c == ' ') {
                                    '-';
                                  } else {
                                    c;
                                  }
                                )

                           | _ => testStatus.url
                           }
                         )
                         |> React.string}
                      </td>
                      <td className="border px-4 py-2">
                        {(
                           switch (testStatus.active, testStatus.statusId) {
                           | (false, _)
                           | (true, 5) => "-"
                           | _ =>
                             testStatus.responseTime
                             |> float_of_int
                             |> LogList.milliOfNano
                             |> Js.Float.toFixedWithPrecision(~digits=1)
                           }
                         )
                         |> React.string}
                      </td>
                    </tr>
                  },
                  testsStatus,
                )
                |> React.array}
             )
           </tbody>
         </table>
       }}
    </div>
  </>;
};