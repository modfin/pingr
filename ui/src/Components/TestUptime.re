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

[@react.component]
let make = (~testId) => {
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

  React.useEffect0(() => {
    let cb = result => {
      switch (result) {
      | None => dispatch(LoadFail("Not working"))
      | Some(logs) => dispatch(LoadSuccess(logs))
      };
    };
    Api.fetchLogsWithCallback(~id=testId, ~numDays="1", cb, ());
    None;
  });
  <>
    <div className="relative bg-gray-200 my-1 p-3">
      <p className="text-xl font-bold"> {ReasonReact.string("Uptime")} </p>
    </div>
    {switch (state) {
     | Loading =>
       <div className="h-screen"> {ReasonReact.string("Loading...")} </div>
     | Failed(msg) => <div> {ReasonReact.string(msg)} </div>
     | Success(logs) =>
       let options =
         Highcharts.Options.(
           make(
             ~title=Title.make(~text=Some("Response times last 24H"), ()),
             ~series=
               Models.Log.(
                 [|
                   Series.column(
                     ~data=
                       logs
                       |> List.filter(log => log.statusId != 5)
                       |> List.map(log =>
                            log.responseTime
                            |> float_of_int
                            |> Js.Date.fromFloat
                            |> Js.Date.getMilliseconds
                            |> int_of_float
                          )
                       |> Array.of_list,
                     ~name="Response time",
                     ~colorByPoint=true,
                     ~colors=
                       logs
                       |> List.filter(log => log.statusId != 5)
                       /* red:f45b5b  gr:90ed7d */
                       |> List.map(log =>
                            switch (log.statusId) {
                            | 1 => "#90ed7d" /* green */
                            | 2
                            | 3 => "#f45b5b" /* red */
                            | _ => "#7cb5ec"
                            }
                          )
                       |> Array.of_list,
                     (),
                   ),
                 |]
               ),
             ~xAxis=
               Models.Log.(
                 Axis.make(
                   ~title=Title.make(~text=Some("Time"), ()),
                   ~labels=AxisLabel.make(~enabled=false, ()),
                   ~categories=
                     logs
                     |> List.filter(log => log.statusId != 5)
                     |> List.map(log => log.createdAt)
                     |> Array.of_list,
                   (),
                 )
               ),
             ~yAxis=
               Axis.make(
                 ~title=Title.make(~text=Some("Response time (ms)"), ()),
                 (),
               ),
             (),
           )
         );
       <HighchartsReact options />;
     }}
  </>;
};