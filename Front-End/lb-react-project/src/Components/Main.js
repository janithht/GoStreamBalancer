const myData = [
    {name: 'John'},
    {name: 'Doe'},
    {name: 'Smith'}
  ];
  
  
function Main() {
    return(
        <div>
            Hi {myData.map((data) => {
                return(<>
                    <h1>{data.name}</h1><br />
                </>
                );
            })}
        </div>
    );
}

export default Main;