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
let make = (~logs) => {
  let (viewSuccess, setViewSuccess) = React.useState(() => true);
  let (viewError, setViewError) = React.useState(() => true);
  let (viewTimedOut, setViewTimedOut) = React.useState(() => true);
  let (viewInit, setViewInit) = React.useState(() => true);
  <>
    <div className="-mx-6">
      <Divider title="Logs">
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
            onChange={e => {
              setViewError(ReactEvent.Form.target(e)##checked)
            }}
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
      </Divider>
    </div>
    <table className="w-full text-left mx-2">
      <thead className="flex w-full">
        <tr className="flex w-full">
          <td className="font-bold px-4 py-2 w-1/12">
            {"ID" |> React.string}
          </td>
          <td className="font-bold px-4 py-2 w-1/6 lg:w-1/12">
            {"Status" |> React.string}
          </td>
          <td className="font-bold px-4 py-2 w-1/6">
            {"Response time (ms)" |> React.string}
          </td>
          <td className="font-bold px-4 py-2 w-5/12 lg:w-1/2">
            {"Message" |> React.string}
          </td>
          <td className="font-bold px-4 py-2 w-1/4">
            {"Created at" |> React.string}
          </td>
        </tr>
      </thead>
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
                || log.statusId == 6
              )
           |> List.map(log =>
                <tr key={string_of_int(log.logId)} className="w-full flex">
                  <td className="border px-4 py-2 w-1/12">
                    {log.logId |> string_of_int |> React.string}
                  </td>
                  <td className="border px-4 py-2 w-1/6 lg:w-1/12">
                    {switch (log.statusId) {
                     | 1 =>
                       <p className="text-green-500">
                         {"Success" |> React.string}
                       </p>
                     | 2 =>
                       <p className="text-red-500">
                         {"Error" |> React.string}
                       </p>
                     | 3 =>
                       <p className="text-red-500">
                         {"Timed out" |> React.string}
                       </p>
                     | 5 =>
                       <p className="text-yellow-500">
                         {"Initialized" |> React.string}
                       </p>
                     | 6 =>
                       <p className="text-gray-500">
                         {"Paused" |> React.string}
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
                  <td className="border px-4 py-2 w-5/12 lg:w-1/2">
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
    </table>
    <div className="h-16" />
  </>;
};