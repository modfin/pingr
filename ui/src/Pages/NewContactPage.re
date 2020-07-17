[@react.component]
let make = () => {
  <>
    <div className="relative bg-gray-400 my-4 p-1">
      <p className="text-xl font-bold"> {"New contact" |> React.string} </p>
    </div>
    <ContactForm submitContact=Api.postContact />
  </>;
};