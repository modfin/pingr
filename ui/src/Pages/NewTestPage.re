[@react.component]
let make = () => {
  <>
    <div className="relative bg-gray-400 my-4 p-1">
      <p className="text-xl font-bold"> {"New test" |> React.string} </p>
    </div>
    <TestForm submitTest=Api.postTest submitContacts=Api.postTestContacts />
  </>;
};