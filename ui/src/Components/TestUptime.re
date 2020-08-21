let milliOfNano = (duration): float => {
  duration /. 1000000.;
};

[@react.component]
let make = (~rts, ~statuses, ~times, ~title_) => {
  let options =
    Highcharts.Options.(
      make(
        ~title=Title.make(~text=Some(title_), ()),
        ~series=[|
          Series.column(
            ~data=
              rts
              |> List.rev_map(f =>
                   f
                   |> milliOfNano
                   |> Js.Float.toFixedWithPrecision(~digits=1)
                   |> float_of_string
                 )
              |> Array.of_list,
            ~name="Average response time",
            ~colorByPoint=true,
            ~colors=
              statuses
              /* red:f45b5b  gr:90ed7d */
              |> List.rev_map(s =>
                   switch (s) {
                   | 1 => "#90ed7d" /* green */
                   | 2
                   | 3 => "#f45b5b" /* red */
                   | _ => "#7cb5ec"
                   }
                 )
              |> Array.of_list,
            (),
          ),
        |],
        ~xAxis=
          Axis.make(
            ~title=Title.make(~text=Some("Time"), ()),
            ~labels=AxisLabel.make(~enabled=false, ()),
            ~categories=times |> List.rev_map(t => t) |> Array.of_list,
            (),
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
};