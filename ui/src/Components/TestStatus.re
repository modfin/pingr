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

let milliOfNano = (duration): float => {
  duration /. 1000000.;
};

[@react.component]
let make = (~id) => {
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
      if (state == Loadable.Loading) {
        let cb = result => {
          switch (result) {
          | None => dispatch(LoadFail("Not working"))
          | Some(logs) => dispatch(LoadSuccess(logs))
          };
        };
        Api.fetchLogsWithCallback(~id, ~numDays, cb, ());
      };
      None;
    },
    [|state|],
  );
  <>
    <Divider title="Status">
      <button
        type_="button"
        onClick={_e => {dispatch(LoadData)}}
        className="bg-blue-500 hover:bg-blue-700 text-white py-1 px-2 rounded">
        {ReasonReact.string("Refresh")}
      </button>
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
    </Divider>
    <div className="px-6">
      Models.Log.(
        {switch (state) {
         | Success(logs) =>
           let timespan =
             switch (numDays) {
             | "1" => 15. *. 60000. // 15 minutes
             | "7" => 60. *. 60000. // 1 hour
             | "30" => 240. *. 60000. // 4 hours
             | "365" => 864. *. 100000. // 1 day
             | "0" => 6048. *. 100000. // 1 week
             | _ => 15. *. 60000.
             };
           let prevIndex = ref(0);
           let currentTime =
             ref(Js.Date.parseAsFloat(List.hd(logs).createdAt));
           currentTime := currentTime^ -. mod_float(currentTime^, timespan);
           let times = ref([]);
           let avgResponses = ref([]);
           let statuses = ref([]);
           let logArr =
             logs
             |> List.filter(log =>
                  Models.Log.(log.statusId != 5 && log.statusId != 6)
                )
             |> Array.of_list;
           let logsLength = Array.length(logArr);
           Belt.Array.forEachWithIndex(
             logArr,
             (i, log) => {
               if (Js.Date.parseAsFloat(log.createdAt) < currentTime^) {
                 let len = i - prevIndex^;
                 statuses :=
                   statuses^
                   @ [
                     Belt.Array.reduce(
                       Belt.Array.slice(logArr, ~offset=prevIndex^, ~len),
                       0.,
                       (a, b) => {
                       a +. float_of_int(b.statusId)
                     })
                     /. float_of_int(len),
                   ];
                 times :=
                   times^
                   @ [Js.Date.fromFloat(currentTime^) |> Js.Date.toString];
                 avgResponses :=
                   avgResponses^
                   @ [
                     Belt.Array.reduce(
                       Belt.Array.slice(logArr, ~offset=prevIndex^, ~len),
                       0.,
                       (a, b) => {
                       a +. float_of_int(b.responseTime)
                     })
                     /. float_of_int(len),
                   ];
                 prevIndex := i;
                 currentTime := currentTime^ -. timespan;
               };
               ();
             },
           );
           statuses :=
             statuses^
             @ [
               Belt.Array.reduce(
                 Belt.Array.sliceToEnd(logArr, prevIndex^), 0., (a, b) => {
                 a +. float_of_int(b.statusId)
               })
               /. float_of_int(logsLength - prevIndex^),
             ];
           times :=
             times^ @ [Js.Date.fromFloat(currentTime^) |> Js.Date.toString];
           avgResponses :=
             avgResponses^
             @ [
               Belt.Array.reduce(
                 Belt.Array.sliceToEnd(logArr, prevIndex^), 0., (a, b) => {
                 a +. float_of_int(b.responseTime)
               })
               /. float_of_int(logsLength - prevIndex^),
             ];
           <>
             <TestUptime
               title_="Average response times"
               rts=avgResponses^
               times=times^
               statuses={statuses^ |> List.map(s => s <= 1. ? 1 : 2)}
             />
             <LogList logs />
           </>;

         | Loading =>
           <div className="h-screen"> {"Loading..." |> React.string} </div>
         | _ => "Error" |> React.string
         }}
      )
    </div>
  </>;
};