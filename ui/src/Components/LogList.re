module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadData
  | LoadSuccess(list(Models.Log.t))
  | LoadFail(string);

type testState =
  | NotAsked
  | Loading
  | Success(list(Models.Log.t))
  | Failure;

type state = Loadable.t(list(Models.Log.t));

let toMicroSeconds = (duration): float => {
  duration /. 100000.;
};

[@react.component]
let make = (~testId=?, ()) => {
  let (viewSuccess, setViewSuccess) = React.useState(() => true);
  let (viewError, setViewError) = React.useState(() => true);
  let (viewTimedOut, setViewTimedOut) = React.useState(() => true);
  let (viewInit, setViewInit) = React.useState(() => true);
  let (numDays, setNumDays) = React.useState(() => "1");

  let (state, dispatch) =
    React.useReducer(
      (_state, action) =>
        switch (action) {
        | LoadData => Loadable.Loading
        | LoadSuccess(posts) => Loadable.Success(posts)
        | LoadFail(msg) => Loadable.Failed(msg)
        },
      Loadable.Loading,
    );

  React.useEffect1(
    () => {
      switch (state) {
      | Loadable.Loading =>
        let cb = result => {
          switch (result) {
          | None => dispatch(LoadFail("Not working"))
          | Some(logs) => dispatch(LoadSuccess(logs))
          };
        };
        switch (testId) {
        | None => Api.fetchLogsWithCallback(~numDays, cb, ())
        | Some(id) => Api.fetchLogsWithCallback(~id, ~numDays, cb, ())
        };
      | _ => ()
      };
      None;
    },
    [|state|],
  );
  <>
    <div className="relative bg-gray-200 my-1 p-2">
      <p className="text-xl font-bold"> {ReasonReact.string("Logs")} </p>
      <button
        type_="button"
        onClick={_e => {dispatch(LoadData)}}
        className="bg-blue-500 hover:bg-blue-700 text-white py-1 px-2 rounded">
        <p className="text-sm"> {ReasonReact.string("Refresh")} </p>
      </button>
      <p className="text-lg mt-1"> {"Status filter " |> React.string} </p>
      <label className="test-gray-700 mr-2">
        <input
          type_="checkbox"
          className="align-middle"
          checked=viewSuccess
          onChange={e => {
            setViewSuccess(ReactEvent.Form.target(e)##checked)
          }}
        />
        {" Success" |> React.string}
      </label>
      <label className="test-gray-700 m-2">
        <input
          type_="checkbox"
          className="align-middle"
          checked=viewError
          onChange={e => {setViewError(ReactEvent.Form.target(e)##checked)}}
        />
        {" Error" |> React.string}
      </label>
      <label className="test-gray-700 m-2 text-center">
        <input
          type_="checkbox"
          className="align-middle"
          checked=viewTimedOut
          onChange={e => {
            setViewTimedOut(ReactEvent.Form.target(e)##checked)
          }}
        />
        {" Timed out" |> React.string}
      </label>
      <label className="test-gray-700 m-2 text-center">
        <input
          type_="checkbox"
          className="align-middle"
          checked=viewInit
          onChange={e => {setViewInit(ReactEvent.Form.target(e)##checked)}}
        />
        {" Initialized" |> React.string}
      </label>
      <p className="text-lg mt-1"> {"Since" |> React.string} </p>
      <label className="test-gray-700 text-center">
        <select
          onChange={e => {
            setNumDays(ReactEvent.Form.target(e)##value);
            dispatch(LoadData);
          }}
          name="time"
          className="align-middle">
          <option value="1"> {"Last day" |> React.string} </option>
          <option value="7"> {"Last week" |> React.string} </option>
          <option value="30"> {"Last month" |> React.string} </option>
          <option value="365"> {"Last year" |> React.string} </option>
          <option value="0"> {"All" |> React.string} </option>
        </select>
      </label>
    </div>
    <table className="w-full text-left mx-2">
      <thead className="flex w-full">
        <tr className="flex w-full">
          <td className="font-bold px-4 py-2 w-1/6">
            {"ID" |> React.string}
          </td>
          <td className="font-bold px-4 py-2 w-1/6">
            {"Status" |> React.string}
          </td>
          <td className="font-bold px-4 py-2 w-1/6">
            {"Response time (ms)" |> React.string}
          </td>
          <td className="font-bold px-4 py-2 w-1/4">
            {"Message" |> React.string}
          </td>
          <td className="font-bold px-4 py-2 w-1/4">
            {"Created at" |> React.string}
          </td>
        </tr>
      </thead>
      {switch (state) {
       | Loading =>
         <div className="h-screen"> {ReasonReact.string("Loading...")} </div>
       | Failed(msg) => <div> {ReasonReact.string(msg)} </div>
       | Success(logs) =>
         <tbody
           className="flex flex-col justify-between items-center overflow-y-scroll max-h-screen w-full">
           Models.Log.(
             {logs
              |> List.filter(log =>
                   viewSuccess
                   && log.statusId == 1
                   || viewError
                   && log.statusId == 2
                   || viewTimedOut
                   && log.statusId == 3
                   || viewInit
                   && log.statusId == 5
                 )
              |> List.map(log =>
                   <tr key={string_of_int(log.logId)} className="w-full flex">
                     <td className="border px-4 py-2 w-1/6">
                       {log.logId |> string_of_int |> React.string}
                     </td>
                     <td className="border px-4 py-2 w-1/6">
                       {switch (log.statusId) {
                        | 1 =>
                          <p className="text-green-600">
                            {"Success" |> React.string}
                          </p>
                        | 2 =>
                          <p className="text-red-600">
                            {"Error" |> React.string}
                          </p>
                        | 3 =>
                          <p className="text-red-600">
                            {"Timed out" |> React.string}
                          </p>
                        | 5 =>
                          <p className="text-gray-600">
                            {"Initialized" |> React.string}
                          </p>
                        | _ => "Unknown status id" |> React.string
                        }}
                     </td>
                     <td className="border px-4 py-2 w-1/6">
                       {React.string(
                          log.responseTime
                          |> float_of_int
                          |> toMicroSeconds
                          |> Js.Float.toFixedWithPrecision(~digits=1),
                        )}
                     </td>
                     <td className="border px-4 py-2 w-1/4">
                       {log.message |> React.string}
                     </td>
                     <td className="border px-4 py-2 w-1/4">
                       {log.createdAt
                        |> Js.Date.fromString
                        |> Js.Date.toLocaleString
                        |> React.string}
                     </td>
                   </tr>
                 )
              |> Array.of_list
              |> ReasonReact.array}
           )
         </tbody>
       }}
    </table>
  </>;
};