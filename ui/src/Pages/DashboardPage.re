let options =
  Highcharts.Options.(
    make(
      ~title=Title.make(~text=Some("Response times"), ()),
      ~series=[|
        Series.column(~data=[|1., 2., 3., 4.|], ~name="Response time", ()),
      |],
      ~xAxis=
        Axis.make(
          /*~labels=AxisLabel.make(~align=`center, ~format="value", ~x=1, ()),*/
          ~categories=[|"a", "b", "c"|],
          (),
        ),
      (),
    )
  );

[@react.component]
let make = () => {
  <HighchartsReact options />;
};