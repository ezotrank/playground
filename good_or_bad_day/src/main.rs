use std::env;

use futures::StreamExt;

use telegram_bot::{prelude::*};
use telegram_bot::{Api, MessageKind, UpdateKind, ChatId, PollAnswer, User};

use tokio::time;
use tokio_postgres;

async fn questioner(api: Api, chat_id: i64) -> Result<(), telegram_bot::Error> {
    let start = time::Instant::now();
    let mut interval = time::interval_at(start, time::Duration::from_secs(3600 * 24)); 
    loop {
        tokio::select! {
            _ = interval.tick() => {
            let chat = ChatId::new(chat_id);
            let question = "How's going?";
            let options = vec!["Good", "Bad"];
            let poll = chat.poll(question, options).not_anonymous().to_owned();
            let poll_message = api.send(poll).await?;    
            println!("poll message is {:#?}", poll_message);
            }
        };
    };
    
    Ok(())
}

async fn streamer(api: Api) -> Result<(), telegram_bot::Error> {   
    let pg_dsn = env::var("POSTGRES_DSN").expect("POSTGRES_DSN not set");
    let mut stream = api.stream();
    let (pg_client, conn) = match tokio_postgres::connect(&pg_dsn, tokio_postgres::NoTls).await {
        Ok(result) => (result.0, result.1),
        Err(err) => panic!("pg error: {}", err),
    };
    
    
    tokio::spawn(async move {
        if let Err(e) = conn.await {
            eprintln!("connection error: {}", e);
        }
    });    
        
    while let Some(update) = stream.next().await {
        let update = update?;
        
        match update.kind {
            UpdateKind::Message(message) => match message.kind {
                MessageKind::Text { ref data, .. } => {
                    println!("message is {:#?}", message);
                },
                _ => (),
            },
            UpdateKind::PollAnswer(PollAnswer {
                poll_id,
                user: User { id, ..},
                option_ids,
            }) => {
                let result = match option_ids[0] {
                    0 => String::from("good").to_string(),
                    _ => String::from("bad").to_string(),
                };
                match pg_client.execute(
                    "INSERT INTO answers(answer, timezone, local_timestamp, created_at) VALUES ($1, 'Asia/Bangkok', now() at time zone 'Asia/Bangkok', now())",
                    &[&result],
                ).await {
                    Err(err) => panic!("pg insert error: {}", err),
                    other => println!("goo exec? {:#?}", other),
                };
                println!("poll {} {} {:?}", poll_id, id, option_ids)
            },
            x => println!("undefined {:#?}", x),
        }        
    }
    
    Ok(())
}

#[tokio::main]
async fn main() -> Result<(), telegram_bot::Error> {
    let token = env::var("TELEGRAM_BOT_TOKEN").expect("TELEGRAM_BOT_TOKEN not set");
    let chat_id: i64 = env::var("CHAT_ID").expect("CHAT_ID not set").parse().unwrap();

    let api = Api::new(token);

    let api1 = api.clone();
    let fq = questioner(api1, chat_id);
    
    let api2 = api.clone();
    let fs = streamer(api2);
    
    let (_,_) = tokio::join!(fq, fs);
    
    Ok(())
 }

